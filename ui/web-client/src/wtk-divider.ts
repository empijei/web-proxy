import { LitElement, html, css, PropertyValues } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { styleMap } from 'lit/directives/style-map.js';

@customElement('wtk-divider')
export class WTKDivider extends LitElement {
  static styles = css`
    :host {
      display: block;
      width: 100%;
      height: 100%;
      overflow: hidden;
      --gutter-size: 6px;
      --gutter-color: #e0e0e0;
      --gutter-hover: #2196f3;
    }

    .container {
      display: flex;
      width: 100%;
      height: 100%;
    }

    .container.vertical {
      flex-direction: row;
    }

    .container.horizontal {
      flex-direction: column;
    }

    .slot-wrapper {
      overflow: auto;
      position: relative;
    }

    .gutter {
      background-color: var(--gutter-color);
      flex: 0 0 var(--gutter-size);
      z-index: 10;
      transition: background-color 0.2s ease;
    }

    .gutter:hover,
    .gutter.dragging {
      background-color: var(--gutter-hover);
    }

    .vertical > .gutter {
      cursor: col-resize;
    }

    .horizontal > .gutter {
      cursor: row-resize;
    }
  `;

  /** Direction of the split: 'vertical' (side-by-side) or 'horizontal' (stacked) */
  @property({ type: String }) direction: 'vertical' | 'horizontal' = 'vertical';

  /** The percentage of space the primary (first) slot occupies */
  @state() private percentage: number = 50;

  @state() private isDragging: boolean = false;

  protected render() {
    const primaryStyle = {
      flex: `0 0 ${this.percentage}%`,
    };

    return html`
      <div class="container ${this.direction}">
        <div class="slot-wrapper" style=${styleMap(primaryStyle)}>
          <slot name="primary"></slot>
        </div>

        <div
          class="gutter ${this.isDragging ? 'dragging' : ''}"
          @mousedown=${this.startDragging}
        ></div>

        <div class="slot-wrapper" style="flex: 1 1 0%;">
          <slot name="secondary"></slot>
        </div>
      </div>
    `;
  }

  private startDragging = (e: MouseEvent): void => {
    e.preventDefault();
    this.isDragging = true;

    window.addEventListener('mousemove', this.onDrag);
    window.addEventListener('mouseup', this.stopDragging);

    // Prevent text selection and force cursor style across the UI
    document.body.style.cursor =
      this.direction === 'vertical' ? 'col-resize' : 'row-resize';
    document.body.style.userSelect = 'none';
  };

  private onDrag = (e: MouseEvent): void => {
    if (!this.isDragging) return;

    const rect = this.getBoundingClientRect();
    let newPercentage: number;

    if (this.direction === 'vertical') {
      // Calculate based on width (X-axis)
      newPercentage = ((e.clientX - rect.left) / rect.width) * 100;
    } else {
      // Calculate based on height (Y-axis)
      newPercentage = ((e.clientY - rect.top) / rect.height) * 100;
    }

    // Constraints: keep the split between 5% and 95%
    this.percentage = Math.min(Math.max(newPercentage, 5), 95);

    // Optional: Dispatch a custom event for the parent to react to resizing
    this.dispatchEvent(
      new CustomEvent('resizing', {
        detail: { percentage: this.percentage },
        bubbles: true,
        composed: true,
      }),
    );
  };

  private stopDragging = (): void => {
    this.isDragging = false;
    window.removeEventListener('mousemove', this.onDrag);
    window.removeEventListener('mouseup', this.stopDragging);

    document.body.style.cursor = '';
    document.body.style.userSelect = '';
  };
}
