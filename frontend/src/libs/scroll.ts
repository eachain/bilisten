let onScrollCallback = (_: Event) => {};

export function setOnScroll(cb: (event: Event) => void) {
  onScrollCallback = cb;
}

export function onScroll(event: Event) {
  onScrollCallback(event);
}

export let scrollToTop: undefined | (() => void) = undefined;

export function setScrollToTop(cb: () => void) {
  scrollToTop = cb;
}
