import { ReactiveController, ReactiveControllerHost } from 'lit';

export class SSEController<ResponseT, RequestT> implements ReactiveController {
  private readonly url: URL;

  private eventSource?: EventSource;

  public value?: ResponseT;
  public err?: Error;

  constructor(
    private readonly host: ReactiveControllerHost,
    origin: string,
    path: string,
    req: RequestT,
  ) {
    (this.host = host).addController(this);
    this.url = new URL(path, origin || window.location.origin);
    this.url.searchParams.set('srpc', JSON.stringify(req));
  }
  hostConnected() {
    if (this.eventSource) {
      return;
    }
    this.eventSource = new EventSource(this.url);
    this.eventSource.addEventListener('val', event => {
      this.value = JSON.parse(event.data);
    });
    this.eventSource.addEventListener('err', event => {
      this.err = new Error(event.data);
    });
    this.eventSource.onerror = err => {
      err.preventDefault();
      this.err = new Error(err.toString());
    };
  }
  hostDisconnected() {
    if (!this.eventSource) {
      return;
    }
    this.eventSource.close();
    this.eventSource = undefined;
  }
}

export class SSEControllerAccum<
  ResponseT,
  RequestT,
> implements ReactiveController {
  private eventSource?: EventSource;

  public value: ResponseT[] = [];
  public lastErr?: Error;
  private readonly url: URL;

  constructor(
    private readonly host: ReactiveControllerHost,
    origin: string,
    path: string,
    req: RequestT,
  ) {
    (this.host = host).addController(this);
    this.url = new URL(path, origin || window.location.origin);
    this.url.searchParams.set('srpc', JSON.stringify(req));
  }
  hostConnected() {
    if (this.eventSource) {
      return;
    }
    this.eventSource = new EventSource(this.url);
    this.eventSource.addEventListener('val', event => {
      this.value.push(JSON.parse(event.data));
    });
    this.eventSource.addEventListener('err', event => {
      this.lastErr = new Error(event.data);
    });
    this.eventSource.onerror = err => {
      err.preventDefault();
      this.lastErr = new Error(err.toString());
    };
  }
  hostDisconnected() {
    if (!this.eventSource) {
      return;
    }
    this.eventSource.close();
    this.eventSource = undefined;
  }
}
