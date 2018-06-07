// Copyright 2018 The Cockroach Authors.
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
//
// This file implements routines for manipulating filtering expressions.

package sql

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type planNodeStack struct {
	stack []*roachpb.PlanNode
}

func (ns *planNodeStack) push(node *roachpb.PlanNode) {
	ns.stack = append(ns.stack, node)
}

func (ns *planNodeStack) pop() *roachpb.PlanNode {
	if len(ns.stack) == 0 {
		return nil
	}
	stackTop := ns.stack[len(ns.stack)-1]
	ns.stack = ns.stack[0 : len(ns.stack)-1]
	return stackTop
}

func (ns *planNodeStack) peek() *roachpb.PlanNode {
	if len(ns.stack) == 0 {
		return nil
	}
	return ns.stack[len(ns.stack)-1]
}

func (ns *planNodeStack) len() int {
	return len(ns.stack)
}

func getPlanTree(ctx context.Context, top planTop) *roachpb.PlanNode {
	nodeStack := planNodeStack{}

	observer := planObserver{
		enterNode: func(ctx context.Context, nodeName string, plan planNode) (bool, error) {
			nodeStack.push(&roachpb.PlanNode{
				Name: nodeName,
			})
			return true, nil
		},
		expr: func(nodeName, fieldName string, n int, expr tree.Expr) {
			if expr == nil {
				return
			}
			stackTop := nodeStack.peek()
			stackTop.Attrs = append(stackTop.Attrs, &roachpb.PlanNode_Attr{
				Key:   fieldName,
				Value: expr.String(),
			})
		},
		attr: func(nodeName, fieldName, attr string) {
			stackTop := nodeStack.peek()
			stackTop.Attrs = append(stackTop.Attrs, &roachpb.PlanNode_Attr{
				Key:   fieldName,
				Value: attr,
			})
		},
		leaveNode: func(nodeName string, plan planNode) error {
			if nodeStack.len() == 1 {
				return nil
			}
			poppedNode := nodeStack.pop()
			newStackTop := nodeStack.peek()
			newStackTop.Children = append(newStackTop.Children, poppedNode)
			return nil
		},
	}
	// TODO(vilterp): subqueries
	if err := walkPlan(ctx, top.plan, observer); err != nil {
		panic(fmt.Sprintf("error while walking plan to save it to statement stats: %s", err.Error()))
	}
	return nodeStack.peek()
}
