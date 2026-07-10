<template>
  <!-- Inline "Powered by Claude" badge. No external fetch — self-contained SVG. -->
  <svg width="148" height="22" viewBox="0 0 148 22" xmlns="http://www.w3.org/2000/svg"
       style="flex-shrink:0;" aria-label="Powered by Claude">
    <!-- Pill background -->
    <rect width="148" height="22" rx="11" fill="#2A1F1A"/>
    <rect width="148" height="22" rx="11" fill="none" stroke="#D97757" stroke-width="1" stroke-opacity="0.6"/>

    <!-- Anthropic logomark (simplified geometric "A" form) -->
    <!-- Left stroke of A -->
    <polygon points="10,16 13,7 14.5,10.5 12,16" fill="#D97757"/>
    <!-- Right stroke of A -->
    <polygon points="17,16 14,7 12.5,10.5 15,16" fill="#D97757"/>
    <!-- Crossbar -->
    <rect x="11.5" y="11.5" width="4" height="1.5" fill="#D97757"/>

    <!-- Divider -->
    <line x1="24" y1="5" x2="24" y2="17" stroke="#D97757" stroke-width="0.75" stroke-opacity="0.4"/>

    <!-- "Powered by Claude" text -->
    <text x="30" y="15"
          font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif"
          font-size="9.5"
          font-weight="500"
          letter-spacing="0.02em"
          fill="#F0E8E4">
      Powered by
      <tspan font-weight="700" fill="#D97757"> Claude</tspan>
    </text>
  </svg>
</template>
