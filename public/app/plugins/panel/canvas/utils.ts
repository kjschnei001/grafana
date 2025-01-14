import { isNumber, isString } from 'lodash';

import { AppEvents, Field, getFieldDisplayName, LinkModel, PluginState, SelectableValue } from '@grafana/data';
import appEvents from 'app/core/app_events';
import { hasAlphaPanels, config } from 'app/core/config';
import {
  defaultElementItems,
  advancedElementItems,
  CanvasElementItem,
  canvasElementRegistry,
  CanvasElementOptions,
  TextConfig,
  CanvasConnection,
} from 'app/features/canvas';
import { notFoundItem } from 'app/features/canvas/elements/notFound';
import { ElementState } from 'app/features/canvas/runtime/element';
import { FrameState } from 'app/features/canvas/runtime/frame';
import { Scene, SelectionParams } from 'app/features/canvas/runtime/scene';
import { DimensionContext } from 'app/features/dimensions';

import { AnchorPoint, ConnectionState } from './types';

export function doSelect(scene: Scene, element: ElementState | FrameState) {
  try {
    let selection: SelectionParams = { targets: [] };
    if (element instanceof FrameState) {
      const targetElements: HTMLDivElement[] = [];
      targetElements.push(element?.div!);
      selection.targets = targetElements;
      selection.frame = element;
      scene.select(selection);
    } else {
      scene.currentLayer = element.parent;
      selection.targets = [element?.div!];
      scene.select(selection);
    }
  } catch (error) {
    appEvents.emit(AppEvents.alertError, ['Unable to select element, try selecting element in panel instead']);
  }
}

export function getElementTypes(shouldShowAdvancedTypes: boolean | undefined, current?: string): RegistrySelectInfo {
  if (shouldShowAdvancedTypes) {
    return getElementTypesOptions([...defaultElementItems, ...advancedElementItems], current);
  }

  return getElementTypesOptions([...defaultElementItems], current);
}

interface RegistrySelectInfo {
  options: Array<SelectableValue<string>>;
  current: Array<SelectableValue<string>>;
}

export function getElementTypesOptions(items: CanvasElementItem[], current: string | undefined): RegistrySelectInfo {
  const selectables: RegistrySelectInfo = { options: [], current: [] };
  const alpha: Array<SelectableValue<string>> = [];

  for (const item of items) {
    const option: SelectableValue<string> = { label: item.name, value: item.id, description: item.description };
    if (item.state === PluginState.alpha) {
      if (!hasAlphaPanels) {
        continue;
      }
      option.label = `${item.name} (Alpha)`;
      alpha.push(option);
    } else {
      selectables.options.push(option);
    }

    if (item.id === current) {
      selectables.current.push(option);
    }
  }

  for (const a of alpha) {
    selectables.options.push(a);
  }

  return selectables;
}

export function onAddItem(sel: SelectableValue<string>, rootLayer: FrameState | undefined, anchorPoint?: AnchorPoint) {
  const newItem = canvasElementRegistry.getIfExists(sel.value) ?? notFoundItem;
  const newElementOptions: CanvasElementOptions = {
    ...newItem.getNewOptions(),
    type: newItem.id,
    name: '',
  };

  if (anchorPoint) {
    newElementOptions.placement = { ...newElementOptions.placement, top: anchorPoint.y, left: anchorPoint.x };
  }

  if (newItem.defaultSize) {
    newElementOptions.placement = { ...newElementOptions.placement, ...newItem.defaultSize };
  }

  if (rootLayer) {
    const newElement = new ElementState(newItem, newElementOptions, rootLayer);
    newElement.updateData(rootLayer.scene.context);
    rootLayer.elements.push(newElement);
    rootLayer.scene.save();
    rootLayer.reinitializeMoveable();

    setTimeout(() => doSelect(rootLayer.scene, newElement));
  }
}

export function getDataLinks(ctx: DimensionContext, cfg: TextConfig, textData: string | undefined): LinkModel[] {
  const panelData = ctx.getPanelData();
  const frames = panelData?.series;

  const links: Array<LinkModel<Field>> = [];
  const linkLookup = new Set<string>();

  frames?.forEach((frame) => {
    const visibleFields = frame.fields.filter((field) => !Boolean(field.config.custom?.hideFrom?.tooltip));

    if (cfg.text?.field && visibleFields.some((f) => getFieldDisplayName(f, frame) === cfg.text?.field)) {
      const field = visibleFields.filter((field) => getFieldDisplayName(field, frame) === cfg.text?.field)[0];
      if (field?.getLinks) {
        const disp = field.display ? field.display(textData) : { text: `${textData}`, numeric: +textData! };
        field.getLinks({ calculatedValue: disp }).forEach((link) => {
          const key = `${link.title}/${link.href}`;
          if (!linkLookup.has(key)) {
            links.push(link);
            linkLookup.add(key);
          }
        });
      }
    }
  });

  return links;
}

export function isConnectionSource(element: ElementState) {
  return element.options.connections && element.options.connections.length > 0;
}
export function isConnectionTarget(element: ElementState, sceneByName: Map<string, ElementState>) {
  const connections = getConnections(sceneByName);
  return connections.some((connection) => connection.target === element);
}

export function getConnections(sceneByName: Map<string, ElementState>) {
  const connections: ConnectionState[] = [];
  for (let v of sceneByName.values()) {
    if (v.options.connections) {
      v.options.connections.forEach((c, index) => {
        // @TODO Remove after v10.x
        if (isString(c.color)) {
          c.color = { fixed: c.color };
        }

        if (isNumber(c.size)) {
          c.size = { fixed: 2, min: 1, max: 10 };
        }

        const target = c.targetName ? sceneByName.get(c.targetName) : v.parent;
        if (target) {
          connections.push({
            index,
            source: v,
            target,
            info: c,
          });
        }
      });
    }
  }

  return connections;
}

export function getConnectionsByTarget(element: ElementState, scene: Scene) {
  return scene.connections.state.filter((connection) => connection.target === element);
}

export function updateConnectionsForSource(element: ElementState, scene: Scene) {
  const targetConnections = getConnectionsByTarget(element, scene);
  targetConnections.forEach((connection) => {
    const sourceConnections = connection.source.options.connections?.splice(0) ?? [];
    const connections = sourceConnections.filter((con) => con.targetName !== element.getName());
    connection.source.onChange({ ...connection.source.options, connections });
  });
}

export const calculateCoordinates = (
  sourceRect: DOMRect,
  parentRect: DOMRect,
  info: CanvasConnection,
  target: ElementState
) => {
  const sourceHorizontalCenter = sourceRect.left - parentRect.left + sourceRect.width / 2;
  const sourceVerticalCenter = sourceRect.top - parentRect.top + sourceRect.height / 2;

  // Convert from connection coords to DOM coords
  const x1 = sourceHorizontalCenter + (info.source.x * sourceRect.width) / 2;
  const y1 = sourceVerticalCenter - (info.source.y * sourceRect.height) / 2;

  let x2;
  let y2;

  if (info.targetName) {
    const targetRect = target.div?.getBoundingClientRect();

    const targetHorizontalCenter = targetRect!.left - parentRect.left + targetRect!.width / 2;
    const targetVerticalCenter = targetRect!.top - parentRect.top + targetRect!.height / 2;

    x2 = targetHorizontalCenter + (info.target.x * targetRect!.width) / 2;
    y2 = targetVerticalCenter - (info.target.y * targetRect!.height) / 2;
  } else {
    const parentHorizontalCenter = parentRect.width / 2;
    const parentVerticalCenter = parentRect.height / 2;

    x2 = parentHorizontalCenter + (info.target.x * parentRect.width) / 2;
    y2 = parentVerticalCenter - (info.target.y * parentRect.height) / 2;
  }
  return { x1, y1, x2, y2 };
};

// @TODO revisit, currently returning last row index for field
export const getRowIndex = (fieldName: string | undefined, scene: Scene) => {
  if (fieldName) {
    const series = scene.context.getPanelData()?.series[0];
    const field = series?.fields.find((f) => (f.name = fieldName));
    const data = field?.values;
    return data ? data.length - 1 : 0;
  }
  return 0;
};

export const getConnectionStyles = (info: CanvasConnection, scene: Scene, defaultArrowSize: number) => {
  const defaultArrowColor = config.theme2.colors.text.primary;
  const lastRowIndex = getRowIndex(info.size?.field, scene);
  const strokeColor = info.color ? scene.context.getColor(info.color).value() : defaultArrowColor;
  const strokeWidth = info.size ? scene.context.getScale(info.size).get(lastRowIndex) : defaultArrowSize;
  return { strokeColor, strokeWidth };
};
