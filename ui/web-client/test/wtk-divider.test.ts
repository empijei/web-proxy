import { html } from 'lit';
import { fixture, expect, oneEvent } from '@open-wc/testing';

import type { WTKDivider } from '../src/wtk-divider.js';
import '../src/wtk-divider.js';

describe('WTKDivider', () => {
  let element: WTKDivider;

  beforeEach(async () => {
    element = await fixture(html`
      <wtk-divider style="width: 1000px; height: 1000px; display: block;">
        <div slot="primary">Primary Content</div>
        <div slot="secondary">Secondary Content</div>
      </wtk-divider>
    `);
  });

  it('is defined', () => {
    const el = document.createElement('wtk-divider');
    expect(el).to.be.instanceOf(customElements.get('wtk-divider'));
  });

  it('renders with default values', () => {
    expect(element.direction).to.equal('vertical');
    const container = element.shadowRoot!.querySelector('.container');
    expect(container?.classList.contains('vertical')).to.be.true;

    const primarySlotWrapper = element.shadowRoot!.querySelector('.slot-wrapper') as HTMLElement;
    expect(primarySlotWrapper.style.flex).to.equal('0 0 50%');
  });

  it('updates direction when property changes', async () => {
    element.direction = 'horizontal';
    await element.updateComplete;

    const container = element.shadowRoot!.querySelector('.container');
    expect(container?.classList.contains('horizontal')).to.be.true;
    expect(container?.classList.contains('vertical')).to.be.false;
  });

  it('passes the a11y audit', async () => {
    await expect(element).shadowDom.to.be.accessible();
  });

  describe('dragging', () => {
    it('updates percentage when dragging vertically', async () => {
      const gutter = element.shadowRoot!.querySelector('.gutter') as HTMLElement;
      const primarySlotWrapper = element.shadowRoot!.querySelector('.slot-wrapper') as HTMLElement;
      const rect = element.getBoundingClientRect();

      // Start dragging at 50%
      gutter.dispatchEvent(new MouseEvent('mousedown', {
        bubbles: true,
        composed: true,
        clientX: rect.left + rect.width / 2,
      }));

      // Move to 75%
      window.dispatchEvent(new MouseEvent('mousemove', {
        clientX: rect.left + rect.width * 0.75,
      }));

      await element.updateComplete;
      expect(primarySlotWrapper.style.flex).to.equal('0 0 75%');

      // Stop dragging
      window.dispatchEvent(new MouseEvent('mouseup'));
    });

    it('updates percentage when dragging horizontally', async () => {
      element.direction = 'horizontal';
      await element.updateComplete;

      const gutter = element.shadowRoot!.querySelector('.gutter') as HTMLElement;
      const primarySlotWrapper = element.shadowRoot!.querySelector('.slot-wrapper') as HTMLElement;
      const rect = element.getBoundingClientRect();

      // Start dragging at 50%
      gutter.dispatchEvent(new MouseEvent('mousedown', {
        bubbles: true,
        composed: true,
        clientY: rect.top + rect.height / 2,
      }));

      // Move to 25%
      window.dispatchEvent(new MouseEvent('mousemove', {
        clientY: rect.top + rect.height * 0.25,
      }));

      await element.updateComplete;
      expect(primarySlotWrapper.style.flex).to.equal('0 0 25%');

      // Stop dragging
      window.dispatchEvent(new MouseEvent('mouseup'));
    });

    it('respects minimum and maximum boundaries', async () => {
      const gutter = element.shadowRoot!.querySelector('.gutter') as HTMLElement;
      const primarySlotWrapper = element.shadowRoot!.querySelector('.slot-wrapper') as HTMLElement;
      const rect = element.getBoundingClientRect();

      // Drag beyond minimum (5%)
      gutter.dispatchEvent(new MouseEvent('mousedown', { clientX: rect.left + rect.width / 2 }));
      window.dispatchEvent(new MouseEvent('mousemove', { clientX: rect.left + rect.width * 0.01 }));
      await element.updateComplete;
      expect(primarySlotWrapper.style.flex).to.equal('0 0 5%');

      // Drag beyond maximum (95%)
      window.dispatchEvent(new MouseEvent('mousemove', { clientX: rect.left + rect.width * 0.99 }));
      await element.updateComplete;
      expect(primarySlotWrapper.style.flex).to.equal('0 0 95%');

      window.dispatchEvent(new MouseEvent('mouseup'));
    });

    it('dispatches "resizing" event while dragging', async () => {
      const gutter = element.shadowRoot!.querySelector('.gutter') as HTMLElement;
      const rect = element.getBoundingClientRect();

      gutter.dispatchEvent(new MouseEvent('mousedown', { clientX: rect.left + rect.width / 2 }));

      setTimeout(() => {
        window.dispatchEvent(new MouseEvent('mousemove', { clientX: rect.left + rect.width * 0.6 }));
      });

      const { detail } = await oneEvent(element, 'resizing');
      expect(detail.percentage).to.be.closeTo(60, 0.1);

      window.dispatchEvent(new MouseEvent('mouseup'));
    });
  });
});
