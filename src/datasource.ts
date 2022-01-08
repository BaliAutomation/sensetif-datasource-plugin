import { DataSourceInstanceSettings, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv, TemplateSrv } from '@grafana/runtime';
import { SensetifDataSourceOptions, SensetifQuery } from './types';

export class DataSource extends DataSourceWithBackend<SensetifQuery, SensetifDataSourceOptions> {
  constructor(
    instanceSettings: DataSourceInstanceSettings<SensetifDataSourceOptions>,
    private readonly templateSrv: TemplateSrv = getTemplateSrv()
  ) {
    super(instanceSettings);
  }

  applyTemplateVariables(query: SensetifQuery, scopedVars: ScopedVars): Record<string, any> {
    return {
      ...query,
      project: this.templateSrv.replace(query.project, scopedVars),
      subsystem: this.templateSrv.replace(query.subsystem, scopedVars),
      datapoint: this.templateSrv.replace(query.datapoint, scopedVars),
    };
  }
}
