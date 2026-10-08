<script setup lang="ts">
import { HugeiconsIcon } from '@hugeicons/vue';
import { Cancel01Icon } from '@hugeicons/core-free-icons';
import { alerts, cancelAlert } from '../libs/alerts';

</script>

<template>
	<div class="fixed top-24 right-8 z-100 w-64">
		<TransitionGroup name="list" tag="ul">
			<li v-for="alert in alerts" :key="alert.id" class="mb-2">
				<div role="alert" class="alert alert-soft alert-horizontal flex justify-between iterms-center" :class="alert.level">
					<span>{{alert.message}}</span>
					<div>
					<HugeiconsIcon
						@click="cancelAlert(alert.id!)"
						:icon="Cancel01Icon"
						:stroke-width="1"
					/>
					</div>
				</div>
			</li>
		</TransitionGroup>
	</div>
</template>

<style scoped lang="css">
.list-move, /* 对移动中的元素应用的过渡 */
.list-enter-active,
.list-leave-active {
  transition: all 0.3s ease;
}

.list-enter-from,
.list-leave-to {
  opacity: 0;
  transform: translateX(256px);
}

/* 确保将离开的元素从布局流中删除
  以便能够正确地计算移动的动画。 */
.list-leave-active {
  position: absolute;
}
</style>