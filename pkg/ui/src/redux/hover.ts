/**
 * Monitors the currently hovered chart and point in time.
 */

import moment from "moment";
import { Action } from "redux";

import { PayloadAction } from "src/interfaces/action";
import { AdminUIState } from "src/redux/state";

export const HOVER_ON = "cockroachui/hover/HOVER_ON";
export const HOVER_OFF = "cockroachui/hover/HOVER_OFF";

/**
 * HoverInfo is conveys the current hover position to the state.
 */
export interface HoverInfo {
  // The logical hover state.
  hoverChart: string;
  hoverTime: moment.Moment;

  // The physical hover state.
  x: number;
  y: number;
}

export class HoverState {
  // Are we currently hovering over a chart?
  currentlyHovering = false;
  // Which chart are we hovering over?
  hoverChart: string;
  // What point in time are we hovering over?
  hoverTime: moment.Moment;

  // The x-coordinate of the mouse relative to the page.
  x: number;
  // The y-coordinate of the mouse relative to the page.
  y: number;
}

export function hoverReducer(state = new HoverState(), action: Action): HoverState {
  switch (action.type) {
    case HOVER_ON:
      const { payload: hoverInfo } = action as PayloadAction<HoverInfo>;
      return {
        currentlyHovering: true,
        ...hoverInfo,
      };
    case HOVER_OFF:
      return new HoverState();
    default:
      return state;
  }
}

export function hoverOn(hoverInfo: HoverInfo): PayloadAction<HoverInfo> {
  return {
    type: HOVER_ON,
    payload: hoverInfo,
  };
}

export function hoverOff(): Action {
  return {
    type: HOVER_OFF,
  };
}

/**
 * Are we currently hovering, and if so, which chart and when?
 */
export const hoverStateSelector = (state: AdminUIState) => state.hover;
