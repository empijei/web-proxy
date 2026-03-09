import { LitElement, html, css } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';

export interface TabData {
  content: string;
  label: string;
}

@customElement('wtk-tabs')
export class WTKTabs extends LitElement {
  @property({ attribute: false })
  tabs: TabData[] = [];

  @state()
  private activeIndex = 0;

  static styles = css`
    :host {
      display: block;
      font-family: sans-serif;
      border: 1px solid #ddd;
      border-radius: 8px;
      overflow: hidden;
    }

    .tab-bar {
      display: flex;
      background: #f4f4f4;
      border-bottom: 1px solid #ddd;
    }

    .tab-button {
      padding: 12px 20px;
      cursor: pointer;
      border: none;
      background: none;
      outline: none;
      transition: background 0.2s;
      font-weight: bold;
      color: #555;
    }

    .tab-button:hover {
      background: #e0e0e0;
    }

    .tab-button.active {
      background: white;
      color: #000b0f;
      border-bottom: 2px solid #000b0f;
    }

    .tab-content {
      padding: 20px;
      animation: fadeIn 0.1s ease;
    }

    @keyframes fadeIn {
      from {
        opacity: 0;
      }
      to {
        opacity: 1;
      }
    }
  `;

  render() {
    if (!this.tabs || this.tabs.length === 0) {
      return html`<p>No tabs available.</p>`;
    }

    return html`
      <div class="tab-bar" role="tablist">
        ${this.tabs.map(
          (tab, index) => html`
            <button
              class="tab-button ${this.activeIndex === index ? 'active' : ''}"
              role="tab"
              aria-selected="${this.activeIndex === index}"
              @click="${() => (this.activeIndex = index)}"
            >
              ${tab.label}
            </button>
          `,
        )}
      </div>

      <div class="tab-content" role="tabpanel">
        ${this.tabs[this.activeIndex]?.content}
      </div>
    `;
  }
}
