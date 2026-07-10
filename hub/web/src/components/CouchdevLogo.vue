<template>
  <!-- Pixel art: Claude-like figure sitting on a couch. 32×24 grid, each cell = 2px -->
  <svg width="64" height="48" viewBox="0 0 32 24" xmlns="http://www.w3.org/2000/svg"
       style="image-rendering:pixelated;flex-shrink:0;" aria-label="couchdev logo">

    <!-- Couch body (tan/warm brown) -->
    <rect x="2"  y="14" width="28" height="6"  fill="#8B6F47"/>
    <!-- Couch seat cushions (lighter) -->
    <rect x="3"  y="15" width="12" height="4"  fill="#A0825A"/>
    <rect x="17" y="15" width="12" height="4"  fill="#A0825A"/>
    <!-- Couch back -->
    <rect x="2"  y="11" width="28" height="3"  fill="#7A5C38"/>
    <!-- Couch left arm -->
    <rect x="1"  y="11" width="3"  height="9"  fill="#7A5C38"/>
    <!-- Couch right arm -->
    <rect x="28" y="11" width="3"  height="9"  fill="#7A5C38"/>
    <!-- Couch legs -->
    <rect x="3"  y="20" width="2"  height="3"  fill="#5C4223"/>
    <rect x="27" y="20" width="2"  height="3"  fill="#5C4223"/>

    <!-- Figure body (Claude orange) sitting on left cushion -->
    <rect x="7"  y="12" width="6"  height="3"  fill="#D97757"/>
    <!-- Figure head -->
    <rect x="8"  y="9"  width="4"  height="3"  fill="#E8956A"/>
    <!-- Figure hair / top -->
    <rect x="8"  y="8"  width="4"  height="1"  fill="#C45E38"/>
    <!-- Figure legs (dangling) -->
    <rect x="7"  y="15" width="2"  height="3"  fill="#D97757"/>
    <rect x="11" y="15" width="2"  height="3"  fill="#D97757"/>
    <!-- Figure eyes -->
    <rect x="9"  y="10" width="1"  height="1"  fill="#1A1A1A"/>
    <rect x="11" y="10" width="1"  height="1"  fill="#1A1A1A"/>
    <!-- Figure smile -->
    <rect x="9"  y="12" width="3"  height="1"  fill="#C45E38"/>
    <!-- Figure arm resting on couch back -->
    <rect x="6"  y="13" width="2"  height="1"  fill="#D97757"/>

    <!-- Laptop on figure's lap (subtle terminal glow) -->
    <rect x="8"  y="15" width="5"  height="3"  fill="#1E2A1E"/>
    <rect x="8"  y="15" width="5"  height="2"  fill="#2A3D2A"/>
    <!-- Terminal cursor blink -->
    <rect x="9"  y="16" width="1"  height="1"  fill="#4FC08D"/>
  </svg>
</template>
