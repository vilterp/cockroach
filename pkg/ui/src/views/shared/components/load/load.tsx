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

import React from "react";

import Loading from "src/views/shared/components/loading";

interface LoadProps<T> {
  load: () => Promise<T>;
  render: (T) => React.ReactNode;
  renderError?: (Error) => React.ReactNode;
}

enum Status {
  LOADING,
  LOADED,
  ERROR,
}

interface LoadState<T> {
  status: Status;
  error: Error;
  data: T;
}

class Load<T> extends React.Component<LoadProps<T>, LoadState<T>> {

  constructor(props: LoadProps<T>) {
    super(props);
    this.state = {
      status: Status.LOADING,
    };
  }

  componentDidMount() {
    this.props.load().then(
      (value) => {
        this.setState({
          status: Status.LOADED,
          data: value,
        });
      },
      (error) => {
        this.setState({
          status: Status.ERROR,
          error: error,
        });
      },
    );
  }

  render() {
    return (
      <Loading
        loading={this.state.status === Status.LOADING}
        render={() => this.props.render(this.state.data)}
        error={this.state.error}
        renderError={this.props.renderError}
      />
    );
  }

}

export default Load;
