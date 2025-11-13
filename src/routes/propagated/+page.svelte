<script lang="ts">
import { onMount, onDestroy } from 'svelte';
import NDK, { NDKNip07Signer, NDKEvent } from '@nostr-dev-kit/ndk';
import { nip19 } from 'nostr-tools';

    const relayUrls = [
        'wss://relay.damus.io',
        'wss://relay.primal.net',
        //'wss://relay.nostr.band', //This relay sometimes gives a problem.
        'wss://nostr.wine'
    ];

    let ndk:any;
    let subscription:any;
    let contracts:any = []; 
    let isLoading = true;

    function toNpub(hex:any) {
        try {
            return nip19.npubEncode(hex);
        } catch (e) {
            return 'invalid_npub';
        }
    }

    onMount(async () => {
        ndk = new NDK({
            explicitRelayUrls: relayUrls
        });

        await ndk.connect();

        const filter = {
            kinds: [1], 
            '#t': ['mare-nostrum'] 
        };

        subscription = ndk.subscribe(filter, { 
            closeOnEose: false 
        });

        subscription.on('event', (event: NDKEvent) => {
            try {
                const contractData = JSON.parse(event.content);

                if (contractData.spec && contractData.spec.startsWith('simplicity-contract')) {
                    
                    const newContract = {
                        id: event.id,
                        authorNpub: toNpub(event.pubkey),
                        data: contractData
                    };

                    if (!contracts.find(c => c.id === newContract.id)) {
                        contracts = [newContract, ...contracts];
                    }
                }
            } catch (e) {
                console.warn('Recebido evento com tag #simplicity-contract mas JSON inválido:', event.content);
            }
        });

        subscription.on('eose', () => {
            isLoading = false;
        });

        return () => {
            subscription.stop();
            ndk.disconnect();
        };
    });
</script>

<div class="max-w-6xl mx-auto min-h-screen my-auto px-8 py-8">
    <div class="w-full text-center mx-auto mt-12 mb-8">
        <div class="inline-flex gap-2 items-center">
            <img src="/logotransparent.png" alt="MareNostrum logo" class="h-32" />
            <h1 class="text-5xl font-semibold text-white">MareNostrum</h1>
        </div>
    </div>
    <div class="w-full max-w-5xl mx-auto p-4 md:p-6 border rounded-xl bg-white">
        <h2 class="text-xl font-semibold text-gray-700 mb-6">Propagated Simplicity Contracts</h2>

        {#if isLoading}
            <div class="text-center py-10">
                <p class="text-lg text-gray-500">Loading contracts...</p>
                </div>
        {:else if contracts.length === 0}
            <div class="text-center py-10 bg-gray-50 rounded-lg">
                <p class="text-lg text-gray-600">No contract found</p>
                <p class="text-sm text-gray-500">Create and publish a contract to see it propagated in the network.</p>
            </div>
        {:else}
            <div class="shadow overflow-hidden border-b border-gray-200 sm:rounded-lg">
                <table class="min-w-full divide-y divide-gray-200">
                    <thead class="bg-gray-50">
                        <tr>
                            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                Contract
                            </th>
                            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                Address (Testnet)
                            </th>
                            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                Creator (Nostr)
                            </th>
                            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                
                            </th>
                            <th scope="col" class="relative px-6 py-3">
                                <span class="sr-only">Ações</span>
                            </th>
                        </tr>
                    </thead>
                    <tbody class="bg-white divide-y divide-gray-200">
                        {#each contracts as contract (contract.id)}
                            <tr>
                                <td class="px-6 py-4 whitespace-nowrap">
                                    <div class="text-sm font-medium text-gray-900">{contract.data.name}</div>
                                    <div class="text-sm text-gray-500 truncate max-w-xs">{contract.data.description}</div>
                                </td>
                                <td class="px-6 py-4 whitespace-nowrap">
                                    <code class="text-sm text-gray-700 font-mono" title={contract.data.contractAddress}>
                                        {contract.data.contractAddress.substring(0, 12)}...{contract.data.contractAddress.substring(contract.data.contractAddress.length - 8)}
                                    </code>
                                </td>
                                
                                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                    <code class="text-sm text-blue-600 font-mono" title={contract.authorNpub}>
                                        {contract.authorNpub.substring(0, 10)}...{contract.authorNpub.substring(contract.authorNpub.length - 4)}
                                    </code>
                                </td>
                                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                    <button class="text-indigo-600 hover:text-indigo-900">
                                        Claim
                                    </button>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        {/if}
    </div>
</div>