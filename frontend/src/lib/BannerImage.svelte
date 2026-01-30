<script>
  /** @type {string} */
  export let src;
  /** @type {string} */
  export let alt = '';
  /** @type {'sm' | 'md' | 'lg'} */
  export let size = 'md';
  /** @type {string} */
  export let wrapperClass = '';

  const sizeClasses = {
    sm: 'h-28 sm:h-36',
    md: 'h-32 sm:h-40',
    lg: 'h-48'
  };
</script>

<div
  class="w-full overflow-hidden rounded-lg bg-white {sizeClasses[size]} {wrapperClass}"
>
  <img
    src={src}
    alt={alt}
    class="banner-img w-full h-full object-contain object-center"
  />
</div>
