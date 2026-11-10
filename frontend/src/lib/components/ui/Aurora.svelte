<!--
  The aurora behind every page (Sir, 10 Nov 2026: animated, but light on the
  visitor's device). Three soft ribbons drawn with gradients only: no filter,
  no canvas, no script. Each moves by transform alone, so the compositor runs
  it without repainting; the layer is fixed, contained and ignores the
  pointer. Reduced motion holds it still; print drops it.
-->
<div class="aurora" aria-hidden="true">
  <span class="r r1"></span>
  <span class="r r2"></span>
  <span class="r r3"></span>
</div>

<style>
  .aurora {
    position: fixed;
    inset: 0;
    z-index: -1;
    overflow: hidden;
    pointer-events: none;
    contain: strict;
  }
  .r {
    position: absolute;
    border-radius: 50%;
    will-change: transform;
    opacity: var(--aur-o);
  }
  /* Wide, tilted ellipses read as ribbons rather than blobs. */
  .r1 {
    left: -25vmax;
    top: -28vmax;
    width: 95vmax;
    height: 48vmax;
    background: radial-gradient(closest-side, var(--aur-1), transparent 72%);
    transform: rotate(-12deg);
    animation: drift-1 28s var(--ease-in-out) infinite alternate;
  }
  .r2 {
    right: -30vmax;
    top: -12vmax;
    width: 85vmax;
    height: 40vmax;
    background: radial-gradient(closest-side, var(--aur-2), transparent 70%);
    transform: rotate(18deg);
    animation: drift-2 34s var(--ease-in-out) infinite alternate;
  }
  .r3 {
    left: 5vmax;
    top: 42vh;
    width: 80vmax;
    height: 36vmax;
    background: radial-gradient(closest-side, var(--aur-3), transparent 70%);
    transform: rotate(-6deg);
    animation: drift-3 40s var(--ease-in-out) infinite alternate;
  }
  @keyframes drift-1 {
    to { transform: translate3d(14vmax, 8vmax, 0) rotate(4deg) scale(1.12); }
  }
  @keyframes drift-2 {
    to { transform: translate3d(-16vmax, 10vmax, 0) rotate(-6deg) scale(1.08); }
  }
  @keyframes drift-3 {
    to { transform: translate3d(10vmax, -14vmax, 0) rotate(10deg) scale(1.18); }
  }

  @media (prefers-reduced-motion: reduce) {
    .r { animation: none; }
  }
  @media print {
    .aurora { display: none; }
  }
</style>
