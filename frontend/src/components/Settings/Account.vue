<script setup lang="ts">
  import { ref, onActivated } from "vue";
  import { getUserInfo, logout, type UserInfo } from "../../libs/login.ts";
  import { isLogined, proxyImage } from "../../libs/backend.ts";

  const userinfo = ref<UserInfo>({
    isLogin: false,
    face: "",
    uname: "",
    mid: 0n,
  });

  onActivated(() => getUserInfo().then((info) => (userinfo.value = info)));

  const logoutNow = () => {
    logout().then(() =>
      isLogined().then((logined) => {
        if (logined) {
          getUserInfo().then((info) => (userinfo.value = info));
        } else {
          userinfo.value = {
            isLogin: false,
            face: "",
            uname: "",
            mid: 0n,
          };
        }
      }),
    );
  };

  const loginNow = () => {
    isLogined().then(logined => {
      if (logined) {
        getUserInfo().then(info => userinfo.value = info);
      } else {
        window.location.hash = "#/login";
      }
    })
  };
</script>

<template>
  <div v-if="userinfo.isLogin" class="flex items-center gap-4">
    <div class="avatar">
      <div class="w-12 rounded-full h-full flex items-center justify-center">
        <img
          v-if="userinfo.face"
          alt="Avatar"
          :src="proxyImage(userinfo.face)"
        />
      </div>
    </div>
    <span>{{ userinfo.uname }}</span>
    <button class="btn btn-error btn-xs opacity-80" @click="logoutNow">
      退出登录
    </button>
  </div>
  <div v-else class="flex items-center gap-4">
    <span>未登录</span>
    <button class="btn btn-error btn-xs opacity-80" @click="loginNow">
      登录
    </button>
  </div>
</template>
