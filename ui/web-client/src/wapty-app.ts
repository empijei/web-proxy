import { LitElement, html, css } from 'lit';
import { property, customElement } from 'lit/decorators.js';

@customElement('wapty-app')
export class WaptyApp extends LitElement {
  static styles = css``;

  render() {
    return html`<h1>Wapty</h1>`;
  }
}
