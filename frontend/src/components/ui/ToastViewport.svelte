<script lang="ts">
  import { Toast } from '@skeletonlabs/skeleton-svelte';
  import { toaster } from '@/stores/ui/toast';
</script>

<Toast.Group {toaster}>
  {#snippet children(toast)}
    <Toast {toast}>
      <Toast.Message>
        <Toast.Title>{toast.title}</Toast.Title>
        {#if toast.description}
          <Toast.Description>{toast.description}</Toast.Description>
        {/if}
      </Toast.Message>
      {#if toast.closable}
        <Toast.CloseTrigger />
      {/if}
    </Toast>
  {/snippet}
</Toast.Group>
