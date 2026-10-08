<script setup lang="ts">
import { onMounted } from 'vue';
import PlayerController from './Player/PlayerController.vue';
import PlayerInfo from './Player/PlayerInfo.vue';
import PlayerMode from './Player/PlayerMode.vue';
import { loadPlayerState } from '../libs/player.ts';

onMounted(() => loadPlayerState())
</script>

<template>
    <div class="bg-base-200 px-8 py-4 footer flex flex-row gap-8 items-center">
        <div class="flex-none">
            <PlayerController />
        </div>

        <div class="grow">
            <PlayerInfo />
        </div>

        <div class="flex-none">
            <PlayerMode />
        </div>
    </div>
</template>