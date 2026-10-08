import { ref } from 'vue';

export type Alert = {
  id?: number; // DON'T SET IT, auto generate
  level: 'alert-info' | 'alert-success' | 'alert-warning' | 'alert-error';
  message: string;
  timeout?: number; // ms, default 3s
};

export const alerts = ref<Alert[]>([]);
let incrId = 0;

export function noticeAlert(alert: Alert): number {
  incrId++;
  alert.id = incrId;
  alerts.value.push(alert);
  if (alerts.value.length > 5) {
    alerts.value.shift();
  }
  setTimeout(
    () => {
      alerts.value = alerts.value.filter((a) => a.id !== alert.id);
    },
    alert.timeout ?? (alert.level === 'alert-error' ? 5000 : 3000),
  );
  return alert.id;
}

export function cancelAlert(id: number) {
  alerts.value = alerts.value.filter((a) => a.id !== id);
}
