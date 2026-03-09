import { LitElement, html, css } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';

export interface Entry {
  ID: number;
  // Index signature allows dynamic access using header strings
  [key: string]: any;
}

@customElement('wtk-table')
export class WTKTable extends LitElement {
  @property({ type: Array }) rows: Entry[] = [];
  @property({ type: Array }) headers: string[] = [];

  // Track the widths of each column in pixels
  @state() private columnWidths: number[] = [];

  private startX = 0;
  private startWidth = 0;
  private resizingColIndex = -1;

  static styles = css`
    :host {
      display: block;
      overflow-x: auto;
      font-family: sans-serif;
    }
    table {
      border-collapse: collapse;
      width: 100%;
      table-layout: fixed; /* Required to enforce explicit column widths */
    }
    th,
    td {
      border: 1px solid #e0e0e0;
      padding: 8px 12px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      text-align: left;
    }
    th {
      position: relative;
      background-color: #f5f5f5;
      user-select: none;
      font-weight: 600;
    }
    .resizer {
      position: absolute;
      top: 0;
      right: 0;
      width: 6px;
      height: 100%;
      cursor: col-resize;
      user-select: none;
      z-index: 1;
    }
    .resizer:hover,
    .resizer.active {
      background-color: #2196f3;
    }
  `;

  // Initialize column widths whenever the headers array changes
  willUpdate(changedProperties: Map<string, any>) {
    if (changedProperties.has('headers')) {
      // Defaulting to 150px per column, adjusting length to match headers
      this.columnWidths = this.headers.map(
        (_, i) => this.columnWidths[i] || 150,
      );
    }
  }

  private _onMouseDown(e: MouseEvent, index: number) {
    this.resizingColIndex = index;
    this.startX = e.pageX;
    this.startWidth = this.columnWidths[index];

    // Attach listeners to the document to catch movements outside the header
    document.addEventListener('mousemove', this._onMouseMove);
    document.addEventListener('mouseup', this._onMouseUp);
  }

  private _onMouseMove = (e: MouseEvent) => {
    if (this.resizingColIndex === -1) return;

    const delta = e.pageX - this.startX;
    // Set a minimum column width of 50px
    const newWidth = Math.max(50, this.startWidth + delta);

    this.columnWidths[this.resizingColIndex] = newWidth;
    this.requestUpdate();
  };

  private _onMouseUp = () => {
    this.resizingColIndex = -1;
    document.removeEventListener('mousemove', this._onMouseMove);
    document.removeEventListener('mouseup', this._onMouseUp);
  };

  render() {
    return html`
      <table>
        <colgroup>
          ${this.columnWidths.map(w => html`<col style="width: ${w}px;" />`)}
        </colgroup>
        <thead>
          <tr>
            ${this.headers.map(
              (header, index) => html`
                <th>
                  ${header}
                  <div
                    class="resizer ${this.resizingColIndex === index
                      ? 'active'
                      : ''}"
                    @mousedown=${(e: MouseEvent) => this._onMouseDown(e, index)}
                  ></div>
                </th>
              `,
            )}
          </tr>
        </thead>
        <tbody>
          ${repeat(
            this.rows,
            row => row.ID, // Uses the guaranteed ID for highly efficient DOM updates
            row => html`
              <tr>
                ${this.headers.map(header => html` <td>${row[header]}</td> `)}
              </tr>
            `,
          )}
        </tbody>
      </table>
    `;
  }
}
