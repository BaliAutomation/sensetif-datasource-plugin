import React, { PureComponent } from 'react';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { Input, Select } from '@grafana/ui';

import { DataSource } from './datasource';
import { defaultQuery, SensetifDataSourceOptions, SensetifQuery } from './types';
import { getBackendSrv } from '@grafana/runtime';
import { defaults } from 'lodash';

export const API_RESOURCES = '/api/plugins/sensetif-datasource/resources/';

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

  request = (path: string, method: string, body: string, waitTime = 0) => {
    let srv = getBackendSrv();
    let request: Promise<any>;
    switch (method) {
      case 'GET':
        request = srv.get(API_RESOURCES + path, body);
        break;
      case 'PUT':
        request = srv.put(API_RESOURCES + path, body);
        break;
      case 'POST':
        request = srv.post(API_RESOURCES + path, body);
        break;
      case 'DELETE':
        request = srv.delete(API_RESOURCES + path);
        break;
    }
    return new Promise<any>((resolve, reject) => {
      request
        .then((r) =>
          setTimeout(() => {
            resolve(r);
          }, waitTime)
        )
        .catch((err) => {
          reject(err);
        });
    });
  };

  loadProjects = (): Promise<any[]> => this.request('_', 'GET', '', 0);

  loadSubsystems = (projectName: string): Promise<any[]> => this.request(projectName + '/_', 'GET', '', 0);

  loadDatapoints = (projectName: string, subsystemName: string) =>
    this.request(projectName + '/' + subsystemName + '/_', 'GET', '', 0);

  onQueryProjectChange = async (name: string) => {
    const { onChange, query } = this.props;

    onChange({ ...query, project: name, subsystem: '', datapoint: '' });

    const subsystems = await this.loadSubsystems(name);
    this.setState({
      subsystems: subsystems,
      datapoints: [],
    });
  };

  onQuerySubsystemChange = async (name: string) => {
    const { onChange, query } = this.props;
    const datapoints = await this.loadDatapoints(query.project, name);

    onChange({ ...query, subsystem: name, datapoint: '' });

    this.setState({
      datapoints: datapoints,
    });
  };

  onQueryDatapointChange = (name: string) => {
    const { onChange, query } = this.props;
    onChange({ ...query, datapoint: name });
  };

  onAliasChange = (value: string) => {
    const { onChange, query } = this.props;
    onChange({ ...query, alias: value });
  };

  projectOptions = (): Array<SelectableValue<string>> =>
    this.options(
      this.state.projects.map((el) => el.name),
      this.props.query.project
    );
  subsystemOptions = (): Array<SelectableValue<string>> =>
    this.options(
      this.state.subsystems.map((el) => el.name),
      this.props.query.subsystem
    );
  datapointOptions = (): Array<SelectableValue<string>> =>
    this.options(
      this.state.datapoints.map((el) => el.name),
      this.props.query.datapoint
    );

  options = (values: string[], alternative: string): Array<SelectableValue<string>> => {
    let result = values;
    if (result.length === 0 && alternative.length !== 0) {
      result = [alternative];
    }

    return result.map((el) => this.selectableValue(el));
  };

  selectableValue = (val: string): SelectableValue<string> => ({ label: val, value: val });

  reloadProjects = async () => {
    const projects = await this.loadProjects();

    this.setState({
      projects: projects,
    });
  };

  reloadSubsystems = async () => {
    const subsystems = await this.loadSubsystems(this.props.query.project);

    this.setState({
      subsystems: subsystems,
    });
  };

  reloadDatapoints = async () => {
    const datapoints = await this.loadDatapoints(this.props.query.project, this.props.query.subsystem);

    this.setState({
      datapoints: datapoints,
    });
  };

  getDefaultQuery = (): Partial<SensetifQuery> => {
    let result = defaultQuery;

    // if some queries already exists, init based on the last configured
    if (this.props.queries!.length > 1) {
      result = this.props.queries![this.props.queries!.length - 2];
    }

    return result;
  };

  render() {
    const defQuery = this.getDefaultQuery();
    const query = defaults(this.props.query, defQuery);
    const { project, subsystem, datapoint, alias } = query;

    const projects = this.projectOptions();
    const subsystems = this.subsystemOptions();
    const datapoints = this.datapointOptions();

    return (
      <div className="gf-form">
        <Select<string>
          value={project.length ? project : null}
          options={projects}
          onChange={(val) => val.value !== project && this.onQueryProjectChange(val.value!)}
          onOpenMenu={() => this.state.projects.length === 0 && this.reloadProjects()}
          placeholder={'The project to be queried'}
        />

        <Select<string>
          value={subsystem.length ? subsystem : null}
          options={subsystems}
          onChange={(val) => val.value !== subsystem && this.onQuerySubsystemChange(val.value!)}
          onOpenMenu={() => this.state.subsystems.length === 0 && this.reloadSubsystems()}
          placeholder={'The Subsystem within the project to be queried'}
        />

        <Select<string>
          value={datapoint.length ? datapoint : null}
          options={datapoints}
          onChange={(val) => val.value !== datapoint && this.onQueryDatapointChange(val.value!)}
          onOpenMenu={() => this.state.datapoints.length === 0 && this.reloadDatapoints()}
          placeholder={'The Datapoint in the Subsystem'}
        />

        <Input value={alias} placeholder="alias" onChange={(e) => this.onAliasChange(e.currentTarget.value)} css="" />
      </div>
    );
  }
}
