<script setup lang="ts">
import { HugeiconsIcon } from '@hugeicons/vue';
import { Search01Icon, Cancel01Icon } from '@hugeicons/core-free-icons';
import { ArrowUpLeft01Icon } from '@hugeicons/core-free-icons';
import { computed, ref, onMounted } from 'vue';

import { search as apiSearch, suggest, currentKeyword, totalResultCount, hasMore } from '../../libs/search.ts';
import type { SuggestResults } from '../../libs/search.ts';
import { searchResults, setNextPage, nextPage } from '../../libs/search.ts';
import { noticeAlert } from '../../libs/alerts.ts';
import { getSearchHistory, logger, setSearchKeyword } from '../../libs/backend.ts';

const searchInput = ref();
const keyword = ref('');
const focus = ref(false);
const mouseInMenu = ref(false);
const suggests = ref<SuggestResults>([]);
let history: string[] = [];

const showSuggest = computed(() => (focus.value || mouseInMenu.value) && suggests.value.length > 0);

onMounted(() => {
    window.addEventListener('keydown', (e) => {
        if (e.metaKey && (e.key === 'k' || e.code === 'KeyK')) {
            searchInput.value?.focus();
		}
	})

    getSearchHistory().then(list => history = list);
})

const search = () => {
    mouseInMenu.value = false;
    searchInput.value?.blur();
    if (keyword.value === '') {
        return;
    }

    window.location.hash = "#/search";
    if (currentKeyword.value == keyword.value) {
        return;
    }

    setSearchKeyword(keyword.value).then(list => history = list)

    hasMore.value = true;

    currentKeyword.value = keyword.value;
    searchResults.value = [];
    let page = 1;
    let numPages = 1;
    let loading = false;
    setNextPage(async () => {
        if (loading || page > numPages) {
            return;
        }
        loading = true;
        try {
            const hits = await apiSearch(currentKeyword.value, page);
            totalResultCount.value = hits.numResults;
            numPages = hits.numPages;
            searchResults.value.push(...hits.result);
            page++;
            hasMore.value = page < numPages;
        } catch(error) {
            noticeAlert({level: 'alert-error', message: `搜索失败：${error}`});
            logger.error(`search "${currentKeyword.value}" page ${page}: ${error}`);
        }
        loading = false;
    })
    nextPage();
}

let lastKeyword = "";
let loadKeyword = "";
const getSuggests = async () => {
    focus.value = true;
    const term = keyword.value;
    if (!term) {
        const ss: SuggestResults = [];
        for (const k of history) {
            ss.push({term: k, value: k, name: k})
        }
        suggests.value = ss;
        return;
    }
    if (lastKeyword === term || loadKeyword === term) {
        return;
    }
    if (!term.includes(lastKeyword)) {
        suggests.value = [];
    }
    loadKeyword = term;
    try {
        const results = await suggest(term);
        if (keyword.value === term) {
            loadKeyword = "";
            lastKeyword = term;
            suggests.value = results;
        }
    } catch {
    }
}

const autoFill = (term: string) => {
    searchInput.value?.focus();
    keyword.value = term;
    getSuggests();
}

const choose = (term: string) => {
    keyword.value = term;
    search();
}

const clearInput = () => {
    keyword.value = '';
    searchInput.value?.focus();
}

</script>

<template>
    <div
        class="dropdown dropdown-bottom dropdown-start w-full wails-no-drag"
        :class="showSuggest ? 'dropdown-open' : 'dropdown-close'"
    >
        <div
            class="input w-full rounded-full"
        >
            <input
                type="text"
                ref="searchInput"
                placeholder="输入关键词或原视频链接"
                v-model="keyword"
                autocomplete="off"
                @focusin.stop="getSuggests"
                @focusout.stop="focus=false"
                @input="getSuggests"
                @keydown.enter="search"
            />
            <HugeiconsIcon
                @click="clearInput"
                :class="keyword ? 'opacity-50 cursor-default' : 'opacity-0'"
                :icon="Cancel01Icon"
                :stroke-width="1"
            />
            <div class="btn btn-circle btn-xs" @click="search">
                <HugeiconsIcon class="opacity-50" :size="16" :icon="Search01Icon" :stroke-width="1" />
            </div>
        </div>

        <ul
            class="menu dropdown-content bg-base-100 rounded-box z-100 w-full p-0 shadow-sm"
            @mouseenter="mouseInMenu=true"
            @mouseleave="mouseInMenu=false"
        >
            <li v-for="term in suggests" class="flex flex-row flex-nowrap items-center px-0 max-w-full">
                <div class="grow px-4 line-clamp-1" @click="choose(term.value)" v-html="term.name">
                </div>
                <div class="px-2" @click="autoFill(term.value)">
                    <HugeiconsIcon
                        class="opacity-50 p-0"
                        :size="20"
                        :icon="ArrowUpLeft01Icon"
                        :stroke-width="1"
                    />
                </div>
            </li>
        </ul>
    </div>
</template>

<style lang="css">
.suggest_high_light {
    color: var(--color-secondary);
    display: inline;
    font-style: normal;
}
.input:focus-within {
    outline: none;
}
</style>