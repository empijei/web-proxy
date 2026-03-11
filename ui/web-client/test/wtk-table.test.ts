import { html } from 'lit';
import { fixture, expect } from '@open-wc/testing';

import type { WTKTable } from '../src/wtk-table.js';
import '../src/wtk-table.js';

describe('WTKTable', () => {
  let element: WTKTable;

  const testTableData = [{ ID: 1, name: 'Item', count: 5, status: 'Active' }];

  const testTableHeaders = [
    { Key: 'ID', Label: '#' },
    { Key: 'name', Label: 'Name' },
    { Key: 'count', Label: 'Amt' },
    { Key: 'status', Label: 'Status' },
  ];

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

  it('renders headers correctly', () => {
    const ths = element.shadowRoot!.querySelectorAll('th');
    expect(ths.length).to.equal(testTableHeaders.length);
    testTableHeaders.forEach((header, i) => {
      expect(ths[i].textContent?.trim()).to.contain(header.Label);
    });
  });

  it('renders rows correctly', () => {
    const trs = element.shadowRoot!.querySelectorAll('tbody tr');
    expect(trs.length).to.equal(testTableData.length);
  });

  it('renders cell content correctly', () => {
    // Note: ensure test data keys match headers exactly (case-sensitive)
    const data = [{ ID: 2, name: 'Test Item', count: 10, status: 'Pending' }];
    element.rows = data;
    element.requestUpdate();

    const cells = element.shadowRoot!.querySelectorAll('tbody td');
    expect(cells[0].textContent?.trim()).to.equal('2');
    expect(cells[1].textContent?.trim()).to.equal('Test Item');
    expect(cells[2].textContent?.trim()).to.equal('10');
    expect(cells[3].textContent?.trim()).to.equal('Pending');
  });

  it('initializes column widths', () => {
    const cols = element.shadowRoot!.querySelectorAll('col');
    expect(cols.length).to.equal(testTableHeaders.length);
    cols.forEach(col => {
      expect(col.style.width).to.equal('150px');
    });
  });

  // GEMINI: The following tests don't pass, but I know for a fact that the component works because I tested it.
  // Fix the tests.

  it('updates column width on resize', async () => {
    const resizer = element.shadowRoot!.querySelector(
      '.resizer',
    ) as HTMLElement;

    const cols = element.shadowRoot!.querySelectorAll('col');
    const initialWidth = cols[0].style.width;

    // Simulate mouse down on the first resizer
    resizer.dispatchEvent(new MouseEvent('mousedown', { pageX: 100 } as any));

    // Simulate mouse move
    document.dispatchEvent(new MouseEvent('mousemove', { pageX: 150 } as any));

    // Re-render
    await element.updateComplete;

    expect(cols[0].style.width).to.equal(`${initialWidth + 50}px`);

    // Simulate mouse up
    document.dispatchEvent(new MouseEvent('mouseup'));
  });

  it('enforces minimum column width', async () => {
    const resizer = element.shadowRoot!.querySelector(
      '.resizer',
    ) as HTMLElement;

    resizer.dispatchEvent(new MouseEvent('mousedown', { pageX: 500 } as any));
    document.dispatchEvent(new MouseEvent('mousemove', { pageX: 100 } as any)); // Delta = -400
    await element.updateComplete;

    const cols = element.shadowRoot!.querySelectorAll('col');
    expect(cols[0].style.width).to.equal('50px');

    document.dispatchEvent(new MouseEvent('mouseup'));
  });
});
