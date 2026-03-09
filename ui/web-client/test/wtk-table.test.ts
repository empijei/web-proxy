import { html } from 'lit';
import { fixture, expect } from '@open-wc/testing';

import type { WTKTable } from '../src/wtk-table.js';
import '../src/wtk-table.js';

describe('WTKTable', () => {
  let element: WTKTable;

  const testTableData = [{ ID: 1, name: 'Item', count: 5, status: 'Active' }];

  const testTableHeaders = ['ID', 'Name', 'Count', 'Status'];

  beforeEach(async () => {
    element = await fixture(html`
      <wtk-table
        .headers="${testTableHeaders}"
        .rows="${testTableData}"
      ></wtk-table>
    `);
  });

  it('is defined', () => {
    const el = document.createElement('wtk-table');
    expect(el).to.be.instanceOf(customElements.get('wtk-table'));
  });

  it('passes the a11y audit', async () => {
    await expect(element).shadowDom.to.be.accessible();
  });
});
