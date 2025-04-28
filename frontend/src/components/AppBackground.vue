<template>
  <div class="video-background">
    <video autoplay muted loop>
      <source src="/video/gameplay.mp4" type="video/mp4">
    </video>
  </div>
</template>

<script setup lang="ts">
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
  filter: brightness(0.4);
  opacity: 0.7;
}
</style>
