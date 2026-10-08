<script setup lang="ts">
const {title} = defineProps({title: String});
</script>

<template>
	<div class="divider"></div>
	<h2 class="mb-4 font-medium">{{ title }}</h2>
	<div class="ps-8">
		<slot></slot>
	</div>
</template>