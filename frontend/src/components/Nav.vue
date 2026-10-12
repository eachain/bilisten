<script setup lang="ts">
import { HugeiconsIcon } from '@hugeicons/vue';
import { Home07Icon } from '@hugeicons/core-free-icons';
import Menu from './Nav/Menu.vue';
import SearchInput from './Nav/SearchInput.vue';

const backToHome = () => window.location.hash = '#/';
</script>

<template>
  <div class="w-full shadow-sm px-8 py-4 bg-base-200 flex flex-row gap-8 justify-between wails-draggable">
    <div></div>
    <div class="grow flex flex-row lg:max-w-2/5 md:max-w-1/2 sm:max-w-2/3 gap-4">
      <span class="btn btn-circle wails-no-drag" @click="backToHome">
        <HugeiconsIcon class="opacity-50" :icon="Home07Icon" :stroke-width="1" />
      </span>

      <SearchInput />
    </div>
    <Menu />
  </div>
</template>
