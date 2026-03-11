import { html } from 'lit';
import { fixture, expect } from '@open-wc/testing';

import type { WTKTabs } from '../src/wtk-tabs.js';
import '../src/wtk-tabs.js';

describe('WTKTabs', () => {
  let element: WTKTabs;

  const testTabsData = [
    { label: 'Profile', content: () => html`User profile details go here.` },
    { label: 'Settings', content: () => html`<b>This is bold</b>` },
    { label: 'Notifications', content: () => html`View your recent alerts.` },
  ];

  beforeEach(async () => {
    element = await fixture(html`
      <wtk-tabs .tabs="${testTabsData}"></wtk-tabs>
    `);
  });

  it('is defined', () => {
    const el = document.createElement('wtk-tabs');
    expect(el).to.be.instanceOf(customElements.get('wtk-tabs'));
  });

  it('renders the correct number of tabs', () => {
    const buttons = element.shadowRoot!.querySelectorAll('.tab-button');
    expect(buttons.length).to.equal(testTabsData.length);

    testTabsData.forEach((tab, index) => {
      expect(buttons[index].textContent?.trim()).to.equal(tab.label);
    });
  });

  it('initially displays the first tab content', () => {
    const content = element.shadowRoot!.querySelector('.tab-content');
    expect(content?.textContent?.trim()).to.equal(testTabsData[0].content);

    const buttons = element.shadowRoot!.querySelectorAll('.tab-button');
    expect(buttons[0].classList.contains('active')).to.be.true;
    expect(buttons[0].getAttribute('aria-selected')).to.equal('true');
  });

  it('switches content when a tab is clicked', async () => {
    const buttons = element.shadowRoot!.querySelectorAll(
      '.tab-button',
    ) as NodeListOf<HTMLButtonElement>;
    const content = element.shadowRoot!.querySelector('.tab-content');

    // Click the second tab
    buttons[1].click();
    await element.updateComplete;

    expect(content?.textContent?.trim()).to.equal(testTabsData[1].content);
    expect(buttons[1].classList.contains('active')).to.be.true;
    expect(buttons[1].getAttribute('aria-selected')).to.equal('true');
    expect(buttons[0].classList.contains('active')).to.be.false;
    expect(buttons[0].getAttribute('aria-selected')).to.equal('false');

    // Click the third tab
    buttons[2].click();
    await element.updateComplete;

    expect(content?.textContent?.trim()).to.equal(testTabsData[2].content);
    expect(buttons[2].classList.contains('active')).to.be.true;
  });

  it('displays a message when no tabs are provided', async () => {
    element.tabs = [];
    await element.updateComplete;

    const message = element.shadowRoot!.querySelector('p');
    expect(message?.textContent).to.equal('No tabs available.');
  });

  it('passes the a11y audit', async () => {
    await expect(element).shadowDom.to.be.accessible();
  });
});
