---
layout: page
pageClass: im-home
title: W7Panel - 云原生应用管理平台
---

<script lang="ts" setup>
import {withBase} from 'vitepress'
</script>

<section class="text-center flex flex-col flex-1 px-4 md:px-12">
  <div class="flex-1 flex flex-col items-center justify-center space-y-8">
    <div class="flex justify-center">
      <h1
        class="tagline md:py-12 text-center text-4xl md:text-7xl xl:text-8xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-blue-600 to-green-400 dark:from-green-400 dark:to-yellow-500"
      >
        微擎面板，您的专属服务器管家
      </h1>
    </div>
    <p class="w-56 md:w-auto py-4 md:py-3 md:text-2xl">
      每一个细节，都经过精心打磨，只为了提供更好的管理体验。
    </p>
    <div
      class="space-y-2 md:space-y-0 xl:flex justify-center"
    >
      <div
        class="hidden xl:block mr-4 items-center space-around text-gray-700 bg-gray-100 border-0 py-2 px-6 focus:outline-none hover:bg-gray-200 rounded lg:text-lg"
      >
        <code
          class="bash-composer text-gray-700 bg-transparent flex items-center"
        >
          curl -sfL https://cdn.w7.cc/w7panel/install.sh | sh -
        </code>
      </div>
      <a
        :href="withBase('/user-guide/')"
        class="inline-flex items-center space-around text-white bg-indigo-500 border-0 py-2 px-8 focus:outline-none hover:bg-indigo-600 rounded text-lg"
      >
        <span>立即开始</span>
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="h-6 w-6 ml-2"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M17 8l4 4m0 0l-4 4m4-4H3"
          />
        </svg>
      </a>
    </div>
  </div>
</section>
