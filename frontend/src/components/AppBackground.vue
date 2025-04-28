<template>
  <div class="video-background">
    <video autoplay muted loop playsinline preload="auto" :poster="source.poster" >
      <source :src="source.src" type="video/mp4">
    </video>
  </div>
</template>

<script setup lang="ts">
interface Source {
  src: string;
  poster: string;
}

const sourceArr: Source[] = [
  {
    src: "/video/gameplay.mp4",
    poster: "/image/gameplay.png",
  },
  {
    src: "/video/gameplay1.mp4",
    poster: "/image/gameplay1.png",
  },
  {
    src: "/video/gameplay2.mp4",
    poster: "/image/gameplay2.png",
  }
]

const source = sourceArr[Math.floor(Math.random() * sourceArr.length)]
</script>

<style scoped>
.video-background {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: -1;
  overflow: hidden;
}

.video-background video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  opacity: 0.7;
}
</style>
