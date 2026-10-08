<script setup lang="ts">
import { onMounted } from "vue";
import { themeChange } from "theme-change";
onMounted(() => themeChange(false));
</script>

<template>
	<input type="checkbox" value="dark" data-set-theme class="toggle hidden"/>
</template>