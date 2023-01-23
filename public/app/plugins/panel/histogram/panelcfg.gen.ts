// Code generated - EDITING IS FUTILE. DO NOT EDIT.
//
// Generated by:
//     public/app/plugins/gen.go
// Using jennies:
//     TSTypesJenny
//     PluginTSTypesJenny
//
// Run 'make gen-cue' from repository root to regenerate.

import * as ui from '@grafana/schema';

export const PanelCfgModelVersion = Object.freeze([0, 0]);

export interface PanelOptions extends ui.OptionsWithLegend, ui.OptionsWithTooltip {
  /**
   * Offset buckets by this amount
   */
  bucketOffset?: number;
  /**
   * Size of each bucket
   */
  bucketSize?: number;
  /**
   * Combines multiple series into a single histogram
   */
  combine?: boolean;
}

export const defaultPanelOptions: Partial<PanelOptions> = {
  bucketOffset: 0,
};

export interface PanelFieldConfig extends ui.AxisConfig, ui.HideableFieldConfig {
  /**
   * Controls the fill opacity of the bars.
   */
  fillOpacity?: number;
  /**
   * Set the mode of the gradient fill. Fill gradient is based on the line color. To change the color, use the standard color scheme field option.
   * Gradient appearance is influenced by the Fill opacity setting.
   */
  gradientMode?: ui.GraphGradientMode;
  /**
   * Controls line width of the bars.
   */
  lineWidth?: number;
}

export const defaultPanelFieldConfig: Partial<PanelFieldConfig> = {
  fillOpacity: 80,
  gradientMode: ui.GraphGradientMode.None,
  lineWidth: 1,
};