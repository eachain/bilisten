<script setup lang="ts">
import { onMounted, shallowRef } from 'vue';
import Index from './Index.vue';
import Alerts from './Alerts.vue';

const curr = shallowRef(Index);

const routers = new Map()
    .set('/settings', () => import('./Settings.vue'))
    .set('/search', () => import('./Search.vue'))
    .set('/playlist', () => import('./Playlist.vue'))
    .set('/login', () => import('./Login.vue'));

const cached = new Map();
const mounting = new Map();

const getPath = () => {
    return window.location.hash.slice(1);
}

const loadPage = () => {
    const path = getPath();
    if (!routers.has(path)) {
        curr.value = Index;
        return;
    }
    if (cached.has(path)) {
        curr.value = cached.get(path);
        return;
    }
    if (mounting.has(path)) {
        return;
    }
    mounting.set(path, true);
    routers.get(path)().then((view: any) => {
        cached.set(path, view.default);
        mounting.delete(path);
        if (path === getPath()) {
            curr.value = view.default;
        }
    })
}

window.addEventListener('hashchange', loadPage);
onMounted(loadPage);
</script>

<template>
    <Alerts />
    <KeepAlive>
        <component :is="curr" />
    </KeepAlive>
</template>
