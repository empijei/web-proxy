import { LitElement, html, css } from 'lit';
import { property, customElement } from 'lit/decorators.js';

import './wtk-table.js';
import './wtk-tabs.js';
import './wtk-divider.js';

@customElement('wapty-app')
export class WaptyApp extends LitElement {
  static styles = css``;

  render() {
    const testTableData = [
      { ID: 1, name: 'Item', count: 5, status: 'Active' },
      { ID: 2, name: 'Other', count: 7, status: 'Waiting' },
    ];

    const testTableHeaders = ['ID', 'Name', 'Count', 'Status'];

    return html`<wtk-table
      .headers="${testTableHeaders}"
      .rows="${testTableData}"
    ></wtk-table>`;
  }
}

/*

import { LitElement, html, css } from 'lit';
import { property, customElement } from 'lit/decorators.js';

@customElement('wapty-app')
export class WaptyApp extends LitElement {
  static styles = css``;

  render() {
    return html`<h1>Wapty</h1>`;
  }
}
*/
