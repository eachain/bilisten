<script setup lang="ts">
  import { onMounted } from "vue";
  import { themeChange } from "theme-change";
  import ThemeChanger from "./components/Settings/ThemeChanger.vue";
  import Layout from "./components/Layout.vue";
  import Nav from "./components/Nav.vue";
  import Router from "./components/Router.vue";
  import PlayerBar from "./components/PlayerBar.vue";
  import { autoCheckLoginStatus } from "./libs/login.ts";
  import { listenWailsEvents } from "./libs/wails.ts";

  onMounted(() => {
    themeChange();
    autoCheckLoginStatus();
    listenWailsEvents();
  });
</script>

<template>
  <ThemeChanger />

  <Layout>
    <template #header>
      <Nav />
    </template>
    <template #main>
      <Router />
    </template>
    <template #footer>
      <PlayerBar />
    </template>
  </Layout>
</template>
