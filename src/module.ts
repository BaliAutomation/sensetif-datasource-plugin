import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './ConfigEditor';
import { QueryEditor } from './QueryEditor';
import { SensetifQuery, SensetifDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, SensetifQuery, SensetifDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
