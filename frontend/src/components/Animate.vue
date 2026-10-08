<script setup lang="ts">
const { status } = defineProps({status: String}); // 'loading' | 'success' | 'error'
</script>

<template>
	<Transition name="bounce" mode="out-in">
		<slot name="loading" v-if="status === 'loading'" key="loading"></slot>
		<slot name="success" v-else-if="status === 'success'" key="success" class="text-success"></slot>
		<slot name="error" v-else-if="status === 'error'" key="error" class="text-error"></slot>
		<slot v-else key="comp"></slot>
	</Transition>
</template>

<style scoped>
.bounce-leave-active {
  animation: bounce-out 0.2s;
}
@keyframes bounce-out {
  0% {
    transform: scale(1);
  }
  100% {
    transform: scale(0);
  }
}

.bounce-enter-active {
  animation: bounce-in 0.3s;
}
@keyframes bounce-in {
  0% {
    transform: scale(0);
  }
  50% {
    transform: scale(1.25);
  }
  100% {
    transform: scale(1);
  }
}
</style>