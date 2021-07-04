import { DataSourceInstanceSettings } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { SensetifDataSourceOptions, SensetifQuery } from './types';

export class DataSource extends DataSourceWithBackend<SensetifQuery, SensetifDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<SensetifDataSourceOptions>) {
    super(instanceSettings);
  }

  getQueryDisplayText(query: SensetifQuery) {
    let result = '';
    result += query.project.length && query.project;
    result += query.subsystem.length && `/${query.subsystem}`;
    result += query.datapoint.length && `/${query.datapoint}`;
    return result;
  }
}
