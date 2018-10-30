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

import _ from "lodash";
import React from "react";
import { Helmet } from "react-helmet";
import { connect } from "react-redux";

import * as protos from "src/js/protos";
import { refreshSettings } from "src/redux/apiReducers";
import { CachedDataReducerState } from "src/redux/cachedDataReducer";
import { AdminUIState } from "src/redux/state";
import Loading from "src/views/shared/components/loading";

interface SettingsOwnProps {
  settings: CachedDataReducerState<protos.cockroach.server.serverpb.SettingsResponse>;
  refreshSettings: typeof refreshSettings;
}

type SettingsProps = SettingsOwnProps;

/**
 * Renders the Cluster Settings Report page.
 */
class Settings extends React.Component<SettingsProps, {}> {
  refresh(props = this.props) {
    props.refreshSettings(new protos.cockroach.server.serverpb.SettingsRequest());
  }

  componentWillMount() {
    // Refresh settings query when mounting.
    this.refresh();
  }

  renderTable() {
    if (_.isNil(this.props.settings.data)) {
      return null;
    }

    const { key_values } = this.props.settings.data;

    return (
      <table className="settings-table">
        <thead>
          <tr className="settings-table__row settings-table__row--header">
            <th className="settings-table__cell settings-table__cell--header">Setting</th>
            <th className="settings-table__cell settings-table__cell--header">Value</th>
            <th className="settings-table__cell settings-table__cell--header">Description</th>
          </tr>
        </thead>
        <tbody>
          {
            _.chain(_.keys(key_values))
              .sort()
              .map(key => (
                <tr key={key} className="settings-table__row">
                  <td className="settings-table__cell">{key}</td>
                  <td className="settings-table__cell">{key_values[key].value}</td>
                  <td className="settings-table__cell">{key_values[key].description}</td>
                </tr>
              ))
              .value()
          }
        </tbody>
      </table>
    );
  }

  render() {
    return (
      <div className="section">
        <Helmet>
          <title>Cluster Settings | Debug</title>
        </Helmet>
        <h1>Cluster Settings</h1>
        <h2></h2> {/* TODO(vilterp): why is this here? just for spacing? */}
        <Loading
          loading={!this.props.settings.data}
          error={this.props.settings.lastError}
          render={() => (
            <div>
              <p>Note that some settings have been redacted for security purposes.</p>
              {this.renderTable()}
            </div>
          )}
        />
      </div>
    );
  }
}

function mapStateToProps(state: AdminUIState) {
  return {
    settings: state.cachedData.settings,
  };
}

const actions = {
  refreshSettings,
};

export default connect(mapStateToProps, actions)(Settings);
