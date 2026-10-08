<script setup lang="ts">
import { computed } from 'vue';
import { duration, buffered, progress, setAudioPlayer, seek, playing, play, pause } from '../../libs/player.ts';

let audioPlayer: HTMLMediaElement | undefined = undefined;

setAudioPlayer(() => {
  if (!audioPlayer) {
    audioPlayer = document.getElementById('bilisten-audio-player') as HTMLMediaElement;
  }
  return audioPlayer;
});

const loaded = computed(() => buffered.value * 100 / (duration.value || 1));
const range = computed(() => progress.value * 100 / (duration.value || 1));

const seekTo = (e: InputEvent) => {
  const input = (e.target as HTMLInputElement);
  if (duration.value) {
    seek(duration.value * Number(input.value) / 100);
  } else {
    input.value = "0";
  }
};

let isPlaying = false;

const toggleToPause = () => {
  isPlaying = playing.value;
  pause();
}

const toggleToPlay = () => {
  if (isPlaying) {
    play();
  }
}

</script>

<template>
  <div class="grow max-w-5/6 flex flex-col justify-center text-xl bl-slider">
    <progress class="progress w-full bl-track z-0 bg-base-content/10" value="0" max="100" />
    <progress class="progress w-full bl-track z-1 bl-margin bg-transparent text-base-content/10" :value="loaded" max="100" />
    <progress class="progress w-full bl-track z-2 bl-margin bg-transparent text-secondary/60" :value="range" max="100" />

    <audio id="bilisten-audio-player" controls class="hidden"></audio>
    <input
      type="range"
      class="bl-thumb bl-margin w-full z-10"
      min="0"
      max="100"
      step="any"
      :value="range"
      @input="seekTo"
      @mousedown="toggleToPause"
      @mouseup="toggleToPlay"
      @touchstart.passive="toggleToPause"
      @touchend="toggleToPlay"
    />
	</div>
</template>

<style lang="css" scoped>
.bl-slider {
  --line: 4px;
  --margin: -4px; /* = -1*line */
  --radius: 2px; /* = line/2 */

  padding: 0;
  margin: 0;
}

.bl-track {
  height: var(--line);
  border-radius: var(--radius);
}

.bl-margin {
  margin-top: var(--margin);
}

.bl-thumb {
  height: var(--line);
  background: transparent;
  -webkit-appearance: none;
  appearance: none;
  cursor: pointer;
}

.bl-thumb::-webkit-slider-runnable-track {
  height: var(--line);
  background: transparent;
}

.bl-thumb::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: var(--line);
  height: var(--line);
  transition: transform 0.15s;
}

</style>