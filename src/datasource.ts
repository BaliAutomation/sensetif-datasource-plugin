import { DataSourceInstanceSettings, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { SensetifDataSourceOptions, SensetifQuery } from './types';

export class DataSource extends DataSourceWithBackend<SensetifQuery, SensetifDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<SensetifDataSourceOptions>) {
    super(instanceSettings);
  }

  applyTemplateVariables(query: SensetifQuery, scopedVars: ScopedVars): Record<string, any> {
    let templateSrv = getTemplateSrv();
    return {
      ...query,
      project: templateSrv.replace(query.project, scopedVars),
      subsystem: templateSrv.replace(query.subsystem, scopedVars),
      datapoint: templateSrv.replace(query.datapoint, scopedVars),
    };
  }
}
