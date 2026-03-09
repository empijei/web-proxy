import { html } from 'lit';
import { fixture, expect } from '@open-wc/testing';

import type { WaptyApp } from '../src/wapty-app.js';
import '../src/wapty-app.js';

describe('WaptyApp', () => {
  let element: WaptyApp;
  beforeEach(async () => {
    element = await fixture(html`<wapty-app></wapty-app>`);
  });

  it('passes the a11y audit', async () => {
    await expect(element).shadowDom.to.be.accessible();
  });
});
