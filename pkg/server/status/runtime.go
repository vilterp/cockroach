// Copyright 2015 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package status

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/elastic/gosigar"

	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/shirou/gopsutil/disk"
)

var (
	metaCgoCalls       = metric.Metadata{Name: "sys.cgocalls", Help: "Total number of cgo calls"}
	metaGoroutines     = metric.Metadata{Name: "sys.goroutines", Help: "Current number of goroutines"}
	metaGoAllocBytes   = metric.Metadata{Name: "sys.go.allocbytes", Help: "Current bytes of memory allocated by go"}
	metaGoTotalBytes   = metric.Metadata{Name: "sys.go.totalbytes", Help: "Total bytes of memory allocated by go, but not released"}
	metaCgoAllocBytes  = metric.Metadata{Name: "sys.cgo.allocbytes", Help: "Current bytes of memory allocated by cgo"}
	metaCgoTotalBytes  = metric.Metadata{Name: "sys.cgo.totalbytes", Help: "Total bytes of memory allocated by cgo, but not released"}
	metaGCCount        = metric.Metadata{Name: "sys.gc.count", Help: "Total number of GC runs"}
	metaGCPauseNS      = metric.Metadata{Name: "sys.gc.pause.ns", Help: "Total GC pause in nanoseconds"}
	metaGCPausePercent = metric.Metadata{Name: "sys.gc.pause.percent", Help: "Current GC pause percentage"}
	metaCPUUserNS      = metric.Metadata{Name: "sys.cpu.user.ns", Help: "Total user cpu time in nanoseconds"}
	metaCPUUserPercent = metric.Metadata{Name: "sys.cpu.user.percent", Help: "Current user cpu percentage"}
	metaCPUSysNS       = metric.Metadata{Name: "sys.cpu.sys.ns", Help: "Total system cpu time in nanoseconds"}
	metaCPUSysPercent  = metric.Metadata{Name: "sys.cpu.sys.percent", Help: "Current system cpu percentage"}
	metaRSS            = metric.Metadata{Name: "sys.rss", Help: "Current process RSS"}
	metaFDOpen         = metric.Metadata{Name: "sys.fd.open", Help: "Process open file descriptors"}
	metaFDSoftLimit    = metric.Metadata{Name: "sys.fd.softlimit", Help: "Process open FD soft limit"}
	metaUptime         = metric.Metadata{Name: "sys.uptime", Help: "Process uptime in seconds"}

	metaDiskReadCount = metric.Metadata{Name: "sys.disk.read.count"}
	metaDiskReadTime  = metric.Metadata{Name: "sys.disk.read.time"}
	metaDiskReadBytes = metric.Metadata{Name: "sys.disk.read.bytes"}

	metaDiskWriteCount = metric.Metadata{Name: "sys.disk.write.count"}
	metaDiskWriteTime  = metric.Metadata{Name: "sys.disk.write.time"}
	metaDiskWriteBytes = metric.Metadata{Name: "sys.disk.write.bytes"}

	metaIopsInProgress = metric.Metadata{Name: "sys.disk.iopsinprogress"}
)

// getCgoMemStats is a function that fetches stats for the C++ portion of the code.
// We will not necessarily have implementations for all builds, so check for nil first.
// Returns the following:
// allocated uint: bytes allocated by application
// total     uint: total bytes requested from system
// error           : any issues fetching stats. This should be a warning only.
var getCgoMemStats func(context.Context) (uint, uint, error)

// RuntimeStatSampler is used to periodically sample the runtime environment
// for useful statistics, performing some rudimentary calculations and storing
// the resulting information in a format that can be easily consumed by status
// logging systems.
type RuntimeStatSampler struct {
	clock *hlc.Clock

	startTimeNanos int64
	// The last sampled values of some statistics are kept only to compute
	// derivative statistics.
	lastNow       int64
	lastUtime     int64
	lastStime     int64
	lastPauseTime uint64
	lastCgoCall   int64
	lastNumGC     uint32

	// Only show "not implemented" errors once, we don't need the log spam.
	fdUsageNotImplemented bool

	// Metric gauges maintained by the sampler.
	CgoCalls       *metric.Gauge
	Goroutines     *metric.Gauge
	GoAllocBytes   *metric.Gauge
	GoTotalBytes   *metric.Gauge
	CgoAllocBytes  *metric.Gauge
	CgoTotalBytes  *metric.Gauge
	GcCount        *metric.Gauge
	GcPauseNS      *metric.Gauge
	GcPausePercent *metric.GaugeFloat64
	CPUUserNS      *metric.Gauge
	CPUUserPercent *metric.GaugeFloat64
	CPUSysNS       *metric.Gauge
	CPUSysPercent  *metric.GaugeFloat64
	Rss            *metric.Gauge
	FDOpen         *metric.Gauge
	FDSoftLimit    *metric.Gauge
	Uptime         *metric.Gauge // We use a gauge to be able to call Update.
	BuildTimestamp *metric.Gauge

	// Disk stats
	DiskReadBytes *metric.Gauge
	DiskReadTime  *metric.Gauge
	DiskReadCount *metric.Gauge

	DiskWriteBytes *metric.Gauge
	DiskWriteTime  *metric.Gauge
	DiskWriteCount *metric.Gauge

	IopsInProgress *metric.Gauge
}

// MakeRuntimeStatSampler constructs a new RuntimeStatSampler object.
func MakeRuntimeStatSampler(clock *hlc.Clock) RuntimeStatSampler {
	// Construct the build info metric. It is constant.
	// We first build set the labels on the metadata.
	info := build.GetInfo()
	timestamp, err := info.Timestamp()
	if err != nil {
		// We can't panic here, tests don't have a build timestamp.
		log.Warningf(context.TODO(), "Could not parse build timestamp: %v", err)
	}

	// Build information.
	metaBuildTimestamp := metric.Metadata{Name: "build.timestamp", Help: "Build information"}
	metaBuildTimestamp.AddLabel("tag", info.Tag)
	metaBuildTimestamp.AddLabel("go_version", info.GoVersion)

	buildTimestamp := metric.NewGauge(metaBuildTimestamp)
	buildTimestamp.Update(timestamp)

	return RuntimeStatSampler{
		clock:          clock,
		startTimeNanos: clock.PhysicalNow(),
		CgoCalls:       metric.NewGauge(metaCgoCalls),
		Goroutines:     metric.NewGauge(metaGoroutines),
		GoAllocBytes:   metric.NewGauge(metaGoAllocBytes),
		GoTotalBytes:   metric.NewGauge(metaGoTotalBytes),
		CgoAllocBytes:  metric.NewGauge(metaCgoAllocBytes),
		CgoTotalBytes:  metric.NewGauge(metaCgoTotalBytes),
		GcCount:        metric.NewGauge(metaGCCount),
		GcPauseNS:      metric.NewGauge(metaGCPauseNS),
		GcPausePercent: metric.NewGaugeFloat64(metaGCPausePercent),
		CPUUserNS:      metric.NewGauge(metaCPUUserNS),
		CPUUserPercent: metric.NewGaugeFloat64(metaCPUUserPercent),
		CPUSysNS:       metric.NewGauge(metaCPUSysNS),
		CPUSysPercent:  metric.NewGaugeFloat64(metaCPUSysPercent),
		Rss:            metric.NewGauge(metaRSS),
		FDOpen:         metric.NewGauge(metaFDOpen),
		FDSoftLimit:    metric.NewGauge(metaFDSoftLimit),
		Uptime:         metric.NewGauge(metaUptime),
		BuildTimestamp: buildTimestamp,

		DiskReadTime:  metric.NewGauge(metaDiskReadTime),
		DiskReadCount: metric.NewGauge(metaDiskReadCount),
		DiskReadBytes: metric.NewGauge(metaDiskReadBytes),

		DiskWriteTime:  metric.NewGauge(metaDiskWriteTime),
		DiskWriteCount: metric.NewGauge(metaDiskWriteCount),
		DiskWriteBytes: metric.NewGauge(metaDiskWriteBytes),

		IopsInProgress: metric.NewGauge(metaIopsInProgress),
	}
}

// SampleEnvironment queries the runtime system for various interesting metrics,
// storing the resulting values in the set of metric gauges maintained by
// RuntimeStatSampler. This makes runtime statistics more convenient for
// consumption by the time series and status systems.
//
// This method should be called periodically by a higher level system in order
// to keep runtime statistics current.
func (rsr *RuntimeStatSampler) SampleEnvironment(ctx context.Context) {
	// Record memory and call stats from the runtime package.
	// TODO(mrtracy): memory statistics will not include usage from RocksDB.
	// Determine an appropriate way to compute total memory usage.
	numCgoCall := runtime.NumCgoCall()
	numGoroutine := runtime.NumGoroutine()

	// It might be useful to call ReadMemStats() more often, but it stops the
	// world while collecting stats so shouldn't be called too often.
	// NOTE: the MemStats fields do not get decremented when memory is released,
	// to get accurate numbers, be sure to subtract. eg: ms.Sys - ms.HeapReleased for
	// current memory reserved.
	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	// Retrieve Mem and CPU statistics.
	pid := os.Getpid()
	mem := gosigar.ProcMem{}
	if err := mem.Get(pid); err != nil {
		log.Errorf(ctx, "unable to get mem usage: %v", err)
	}
	cpuTime := gosigar.ProcTime{}
	if err := cpuTime.Get(pid); err != nil {
		log.Errorf(ctx, "unable to get cpu usage: %v", err)
	}

	fds := gosigar.ProcFDUsage{}
	if err := fds.Get(pid); err != nil {
		if _, ok := err.(gosigar.ErrNotImplemented); ok {
			if !rsr.fdUsageNotImplemented {
				rsr.fdUsageNotImplemented = true
				log.Warningf(ctx, "unable to get file descriptor usage (will not try again): %s", err)
			}
		} else {
			log.Errorf(ctx, "unable to get file descriptor usage: %s", err)
		}
	}

	// Time statistics can be compared to the total elapsed time to create a
	// useful percentage of total CPU usage, which would be somewhat less accurate
	// if calculated later using downsampled time series data.
	now := rsr.clock.PhysicalNow()
	dur := float64(now - rsr.lastNow)
	// cpuTime.{User,Sys} are in milliseconds, convert to nanoseconds.
	newUtime := int64(cpuTime.User) * 1e6
	newStime := int64(cpuTime.Sys) * 1e6
	uPerc := float64(newUtime-rsr.lastUtime) / dur
	sPerc := float64(newStime-rsr.lastStime) / dur
	pausePerc := float64(ms.PauseTotalNs-rsr.lastPauseTime) / dur
	rsr.lastNow = now
	rsr.lastUtime = newUtime
	rsr.lastStime = newStime
	rsr.lastPauseTime = ms.PauseTotalNs

	var cgoAllocated, cgoTotal uint
	if getCgoMemStats != nil {
		var err error
		cgoAllocated, cgoTotal, err = getCgoMemStats(ctx)
		if err != nil {
			log.Warningf(ctx, "problem fetching CGO memory stats: %s; CGO stats will be empty.", err)
		}
	}

	goAllocated := ms.Alloc
	goTotal := ms.Sys - ms.HeapReleased

	// Get disk stats.
	disksStats, err := disk.IOCountersWithContext(ctx)
	if err != nil {
		log.Warningf(ctx, "problem fetching disk stats: %s; disk stats will be empty.", err)
	}
	summedDiskStats := sumDiskStats(disksStats)

	rsr.DiskReadBytes.Update(int64(summedDiskStats.ReadBytes))
	rsr.DiskReadCount.Update(int64(summedDiskStats.ReadCount))
	rsr.DiskReadTime.Update(int64(summedDiskStats.ReadTime))

	rsr.DiskWriteBytes.Update(int64(summedDiskStats.WriteBytes))
	rsr.DiskWriteCount.Update(int64(summedDiskStats.WriteCount))
	rsr.DiskWriteTime.Update(int64(summedDiskStats.WriteTime))

	rsr.IopsInProgress.Update(int64(summedDiskStats.IopsInProgress))

	// Log summary of statistics to console.
	cgoRate := float64((numCgoCall-rsr.lastCgoCall)*int64(time.Second)) / dur
	log.Infof(ctx, "runtime stats: %s RSS, %d goroutines, %s/%s/%s GO alloc/idle/total, %s/%s CGO alloc/total, %.2fcgo/sec, %.2f/%.2f %%(u/s)time, %.2f %%gc (%dx)",
		humanize.IBytes(mem.Resident), numGoroutine,
		humanize.IBytes(goAllocated), humanize.IBytes(ms.HeapIdle-ms.HeapReleased), humanize.IBytes(goTotal),
		humanize.IBytes(uint64(cgoAllocated)), humanize.IBytes(uint64(cgoTotal)),
		cgoRate, uPerc, sPerc, pausePerc, ms.NumGC-rsr.lastNumGC)
	if log.V(2) {
		log.Infof(ctx, "memstats: %+v", ms)
	}
	rsr.lastCgoCall = numCgoCall
	rsr.lastNumGC = ms.NumGC

	rsr.CgoCalls.Update(numCgoCall)
	rsr.Goroutines.Update(int64(numGoroutine))
	rsr.GoAllocBytes.Update(int64(goAllocated))
	rsr.GoTotalBytes.Update(int64(goTotal))
	rsr.CgoAllocBytes.Update(int64(cgoAllocated))
	rsr.CgoTotalBytes.Update(int64(cgoTotal))
	rsr.GcCount.Update(int64(ms.NumGC))
	rsr.GcPauseNS.Update(int64(ms.PauseTotalNs))
	rsr.GcPausePercent.Update(pausePerc)
	rsr.CPUUserNS.Update(newUtime)
	rsr.CPUUserPercent.Update(uPerc)
	rsr.CPUSysNS.Update(newStime)
	rsr.CPUSysPercent.Update(sPerc)
	rsr.FDOpen.Update(int64(fds.Open))
	rsr.FDSoftLimit.Update(int64(fds.SoftLimit))
	rsr.Rss.Update(int64(mem.Resident))
	rsr.Uptime.Update((now - rsr.startTimeNanos) / 1e9)
}

func sumDiskStats(disksStats map[string]disk.IOCountersStat) disk.IOCountersStat {
	output := disk.IOCountersStat{}
	for _, stats := range disksStats {
		output.WriteCount += stats.WriteCount
		output.WriteTime += stats.WriteTime
		output.WriteBytes += stats.WriteBytes

		output.ReadCount += stats.ReadCount
		output.ReadTime += stats.ReadTime
		output.ReadBytes += stats.ReadBytes

		output.IopsInProgress += stats.IopsInProgress
	}
	return output
}
