import defaults from 'lodash/defaults';

import React, { PureComponent } from 'react';
import { QueryEditorProps } from '@grafana/data';
import { Select } from '@grafana/ui';

import { DataSource } from './datasource';
import { defaultQuery, SensetifDataSourceOptions, SensetifQuery } from './types';
import { getBackendSrv } from '@grafana/runtime';

type Props = QueryEditorProps<DataSource, SensetifQuery, SensetifDataSourceOptions>;
interface State {
  projects: any[];
  subsystems: any[];
  datapoints: any[];
}

export class QueryEditor extends PureComponent<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      datapoints: [],
      subsystems: [],
      projects: [],
    };
  }

  async componentDidMount() {
    const projects = await this.loadProjects();
    this.setState({
      projects: projects,
    });
  }

  exec = (cmd: any) => getBackendSrv().post('/api/plugins/sensetif-datasource/resources/exec', cmd);

  loadProjects = (): Promise<any[]> => this.exec({ action: 'list', resource: 'project' });
  loadSubsystems = (projectName: string) =>
    this.exec({ action: 'list', resource: 'subsystem', params: { project: projectName } });
  loadDatapoints = (projectName: string, subsystemName: string) =>
    this.exec({ action: 'list', resource: 'datapoint', params: { project: projectName, subsystem: subsystemName } });

  onQueryProjectChange = (project: string) => {
    const { onChange, query } = this.props;
    onChange({ ...query, project: project });
  };
  onQuerySubsystemChange = (subsystem: string) => {
    const { onChange, query } = this.props;
    onChange({ ...query, subsystem: subsystem });
  };
  onQueryDatapointChange = (datapoint: string) => {
    const { onChange, query } = this.props;
    onChange({ ...query, datapoint: datapoint });
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { project, subsystem, datapoint } = query;
    console.log(query);

    return (
      <div className="gf-form">
        <Select
          value={project}
          options={this.state.projects.map((el) => ({ label: el.name, value: el.name }))}
          onChange={(val) => {
            console.log(val);
            this.loadSubsystems(val.value).then((result) => this.setState({ subsystems: result }));
            this.onQueryProjectChange(val.value);
          }}
          placeholder={'The project to be queried'}
        />

        <Select
          value={subsystem}
          options={this.state.subsystems.map((el) => ({ label: el.name, value: el.name }))}
          onChange={(val) => {
            console.log(val);
            this.loadDatapoints(project, val.value).then((result) => this.setState({ datapoints: result }));
            this.onQuerySubsystemChange(val.value);
          }}
          placeholder={'The Subsystem within the project to be queried'}
        />

        <Select
          value={datapoint}
          options={this.state.datapoints.map((el) => ({ label: el.name, value: el.name }))}
          onChange={(val) => {
            console.log(val);
            this.onQueryDatapointChange(val.value);
          }}
          placeholder={'The Datapoint in the Subsystem'}
        />
      </div>
    );
  }
}
