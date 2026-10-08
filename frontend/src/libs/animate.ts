export type AnimateStatus = '' | 'loading' | 'success' | 'error';

export const StatusTimeout = 1000;

export function setAnimateStatus(
  set: (status: AnimateStatus) => void,
  status: AnimateStatus,
) {
  set(status);
  if (status === 'success' || status === 'error') {
    setTimeout(() => set(''), 1000);
  }
}
