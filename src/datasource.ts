import { DataSourceInstanceSettings } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { SensetifDataSourceOptions, SensetifQuery } from './types';

export class DataSource extends DataSourceWithBackend<SensetifQuery, SensetifDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<SensetifDataSourceOptions>) {
    super(instanceSettings);
  }
}
