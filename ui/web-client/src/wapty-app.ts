import { LitElement, html, css } from 'lit';
import { customElement } from 'lit/decorators.js';

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

    const testTableHeaders = [
      { Key: 'ID', Label: '#' },
      { Key: 'name', Label: 'Name' },
      { Key: 'count', Label: 'Amt' },
      { Key: 'status', Label: 'Status' },
    ];

    const testTabsData = [
      {
        label: 'Table',
        content: () =>
          html`<wtk-table
            .headers="${testTableHeaders}"
            .rows="${testTableData}"
          ></wtk-table>`,
      },
      { label: 'Styling', content: () => html`<b>This is bold</b>` },
      {
        label: 'Divider',
        content: () => html`
          <div style="width: 100%; height: 100%;">
            <wtk-divider>
              <div slot="primary">Primary Content</div>
              <div slot="secondary">Secondary Content</div>
            </wtk-divider>
          </div>
        `,
      },
    ];

    return html` <wtk-tabs .tabs="${testTabsData}"></wtk-tabs> `;
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
