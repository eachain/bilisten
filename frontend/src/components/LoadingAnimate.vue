<script setup lang="ts">
  import { HugeiconsIcon } from "@hugeicons/vue";
  import { CheckIcon, Cancel01Icon } from "@hugeicons/core-free-icons";
  import Animate from "./Animate.vue";

  const { status, outSize, iconSize } = defineProps({
    status: String, // 'loading' | 'success' | 'error'
    outSize: String, // size-md
    iconSize: Number,
  });
</script>

<template>
  <Animate :status>
    <template #loading>
      <span class="loading loading-spinner loading-md" :class="outSize"></span>
    </template>

    <template #success>
      <span class="flex justify-center items-center" :class="outSize">
        <HugeiconsIcon
          :icon="CheckIcon"
          :size="iconSize"
          class="text-success"
        />
      </span>
    </template>

    <template #error>
      <span class="flex justify-center items-center" :class="outSize">
        <HugeiconsIcon
          :icon="Cancel01Icon"
          :size="iconSize"
          class="text-error"
        />
      </span>
    </template>

    <slot></slot>
  </Animate>
</template>
