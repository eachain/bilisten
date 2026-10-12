<script setup lang="ts">
import { onScroll } from '../libs/scroll.ts';
</script>

<template>
    <div class="flex flex-col w-svw h-svh divide-y divide-slate-400/[.24]">
        <header>
            <slot name="header"></slot>
        </header>

        <main id="app-main" class="grow overflow-y-auto p-8" @scroll="onScroll">
            <slot name="main"></slot>
        </main>

        <footer>
            <slot name="footer"></slot>
        </footer>
    </div>
</template>