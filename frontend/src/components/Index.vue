<script setup lang="ts">
import NewMusics from './Index/NewMusics.vue';
import HotMusics from './Index/HotMusics.vue';
import Fav from './Index/Fav.vue';
import MainLayout from './MainLayout.vue';

</script>

<template>
	<MainLayout :scroll-to-top="true">
		<div
			class="flex flex-col gap-16"
		>
			<div><NewMusics /></div>
			<div><HotMusics /></div>
			<div><Fav /></div>
		</div>
	</MainLayout>
</template>

