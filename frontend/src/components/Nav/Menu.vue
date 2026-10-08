<script setup lang="ts">
  import { HugeiconsIcon } from "@hugeicons/vue";
  import {
    Login01Icon,
    Logout01Icon,
    Settings01Icon,
    Alien01Icon,
    Astronaut02Icon,
  } from "@hugeicons/core-free-icons";
  import { ref, onMounted } from "vue";
  import { isLogined, proxyImage } from "../../libs/backend.ts";
  import {
    getUserInfo,
    logout,
    registerOnLoginStatusChanged,
  } from "../../libs/login.ts";

  const logined = ref(false);
  const avatar = ref("");
  const username = ref("");

  const checkLoginStatus = () =>
    isLogined().then((ok) => {
      logined.value = ok;
      getUserInfo().then((result) => {
        logined.value = result.isLogin;
        avatar.value = result.face;
        username.value = result.uname;
      });
    });
  
  const checkAvatar = () => {
    if (logined.value && !avatar.value) {
      checkLoginStatus();
    }
  };

  const logoutNow = () => {
    if (document.activeElement) {
      (document.activeElement as HTMLElement).blur();
    }
    logout().then(checkLoginStatus);
  };

  onMounted(() => {
    registerOnLoginStatusChanged("nav-mene-logined", checkLoginStatus);
    checkLoginStatus();

    window.addEventListener("keydown", (e) => {
      if (e.metaKey && (e.key === "," || e.code === "Comma")) {
        window.location.hash = "#/settings";
      }
    });
  });

  const locateTo = (hash: string) => {
    window.location.hash = hash;
    if (document.activeElement) {
      (document.activeElement as HTMLElement).blur();
    }
  }
</script>

<template>
  <div class="dropdown dropdown-end wails-no-drag" @click="checkAvatar">
    <div tabindex="0" role="button" class="btn btn-ghost btn-circle">
      <div class="avatar">
        <div class="w-12 rounded-full h-full flex items-center justify-center">
          <img
            v-if="logined && avatar"
            alt="Avatar"
            :src="proxyImage(avatar)"
          />
          <HugeiconsIcon
            v-else-if="logined"
            :icon="Astronaut02Icon"
            :size="32"
            :stroke-width="1"
          />
          <HugeiconsIcon
            v-else
            :icon="Alien01Icon"
            :size="32"
            :stroke-width="1"
          />
        </div>
      </div>
    </div>
    <ul
      tabindex="-1"
      class="menu menu-sm dropdown-content bg-base-100 rounded-box z-1 mt-3 w-26 p-2 shadow"
    >
      <li>
        <span @click="locateTo('#/settings')">
          <HugeiconsIcon class="h-[1.2em]" :icon="Settings01Icon" />
          设置
        </span>
      </li>

      <li v-if="logined">
        <a @click="logoutNow">
          <HugeiconsIcon class="h-[1.2em]" :icon="Logout01Icon" />
          退出
        </a>
      </li>

      <li v-else>
        <span @click="locateTo('#/login')">
          <HugeiconsIcon class="h-[1.2em]" :icon="Login01Icon" />
          登录
        </span>
      </li>
    </ul>
  </div>
</template>
