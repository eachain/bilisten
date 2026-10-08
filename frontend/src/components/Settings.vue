<script setup lang="ts">
import MainLayout from "./MainLayout.vue";
import Buffer from "./Settings/Buffer.vue";
import Cache from "./Settings/Cache.vue";
import Quality from "./Settings/Quality.vue";
import SettingLayout from "./Settings/SettingLayout.vue";
import ThemeSetting from "./Settings/ThemeSetting.vue";
import Account from "./Settings/Account.vue";
</script>

<template>
<div>
    <MainLayout title="设置" :scroll-to-top="true">
        <SettingLayout title="音质">
            <Quality />
        </SettingLayout>

        <SettingLayout title="播放">
            <Buffer />
        </SettingLayout>

        <SettingLayout title="缓存">
            <Cache />
        </SettingLayout>

        <SettingLayout title="主题">
            <ThemeSetting />
        </SettingLayout>

        <SettingLayout title="账号">
            <Account />
        </SettingLayout>
    </MainLayout>
</div>
</template>
