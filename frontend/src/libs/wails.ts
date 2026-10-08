import { Events, Browser } from '@wailsio/runtime';
import { pause, play, playNext, playPrev, toggle } from './player';

export function listenWailsEvents() {
  Events.On('open-settings', () => {
    window.location.hash = '#/settings';
  });

  Events.On('play-toggle', () => {
    toggle();
  });
  Events.On('play-play', () => {
    play();
  });
  Events.On('play-pause', () => {
    pause();
  });
  Events.On('play-next', () => {
    playNext();
  });
  Events.On('play-prev', () => {
    playPrev();
  });
}

export function openURL(url: string) {
  Browser.OpenURL(url);
}
