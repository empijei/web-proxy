import { html } from 'lit';
import { fixture, expect } from '@open-wc/testing';

import type { WaptyApp } from '../src/wapty-app.js';
import '../src/wapty-app.js';

describe('WaptyApp', () => {
  let element: WaptyApp;
  beforeEach(async () => {
    element = await fixture(html`<wapty-app></wapty-app>`);
  });

  it('renders a h1', () => {
    const h1 = element.shadowRoot!.querySelector('h1')!;
    expect(h1).to.exist;
    expect(h1.textContent).to.equal('Wapty');
  });

  it('passes the a11y audit', async () => {
    await expect(element).shadowDom.to.be.accessible();
  });
});
