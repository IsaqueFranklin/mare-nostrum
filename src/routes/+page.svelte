<script lang="ts">
	import NDK, { NDKNip07Signer, type NDKUserProfile, NDKEvent } from "@nostr-dev-kit/ndk";
	import { browser } from '$app/environment';
	import { onMount } from "svelte";
    import axios from "axios"

    interface ContractResponse {
        message: string;
        blockHeight: number;
        price: number;
        program_hex: string;
        address: string;
        name: string;
        description: string;
    }
	let ndk: NDK | null = null;
	let signer: NDKNip07Signer | null = null;
	let userProfile: NDKUserProfile | null = null;
	let pubkey: string | null = null;
	let error: string | null = null;

    let contractType: boolean = false;
    let addressReturned: ContractResponse | null = null;
    let txid: string | null = null;

    let statusMessage = '';
    let isPublishing = false;
    let published = false;
    let publishedEventJson = '';

	async function connect() {
		try {
			if (!browser) return;

			// creating the NDK instance
			ndk = new NDK({
				explicitRelayUrls: ["wss://relay.damus.io", "wss://nos.lol"], // add any relays you want
			});

			// NIP-07 signer, the extension
			signer = new NDKNip07Signer();
			ndk.signer = signer;

			// connecting to relays
			await ndk.connect(2000);

			// user pubkey from the signer
			const user = await signer.user();
			pubkey = user.pubkey;

			// user profile
			userProfile = await user.fetchProfile();

			console.log("Connected to Nostr!");
			console.log("Pubkey:", pubkey);
			console.log("Profile:", userProfile);
		} catch (e) {
			error = (e as Error).message;
			console.error(e);
		}
	}

    let blockHeight: number | null = null;
	let price: number | null = null;
    let name: string | null = null;
    let description: string | null = null;

	async function submitForm() {
		if (blockHeight === null || price === null) {
			alert("Please fill in both fields.");
			return;
		}

		//const data = { blockHeight, price };

		try {
			const { data } = await axios.post("http://localhost:4000/generate-contract", {
				blockHeight,
				price,
                name,
                description
			});

			console.log("Response:", data);
            if(data){
                addressReturned = data;
            }
		} catch (err) {
			console.error(err);
			alert("Error sending data.");
		}
	}

    async function fundContract() {
        if (addressReturned === null) {
			alert("You have no address to fund.");
			return;
		}

        try {
			const { data } = await axios.post("http://localhost:4000/fund-contract", {
				address: addressReturned.address
			});

			console.log("Response:", data);
            if(data){
                txid = data.txid;
            }
		} catch (err) {
			console.error(err);
			alert("Error sending data.");
		}
    }

    async function publishContract() {
        if (!pubkey || !ndk) {
            statusMessage = 'Please, connect using NIP07.';
            return;
        }

        // Template JSON pré-preenchido para o usuário
    const defaultContractData = {
        "spec": "simplicity-contract-v1",
        "name": addressReturned?.name,
        "description": addressReturned?.description,
        "network": "lbtc-testnet",
        "contractAddress": addressReturned?.address,
        "contractHex": addressReturned?.program_hex,
        "txid": txid,
        "oracleInfo": {
            "oracle_sig_required": "...",
            "required_height": addressReturned?.blockHeight,
            "required_price": addressReturned?.price
        },
        "taproot": {
            "decodeString": "50929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547b5bf5a87e86c63" //This is hardcoded, kinda the most probable decodeString used for taproot merkle proof.
        },
        "witnessFormat": {
            "ORACLE_HEIGHT": {
                "value": "<BlockHeight>",
                "type": "u32"
            },
            "ORACLE_PRICE": {
                "value": "<Price>",
                "type": "u32"
            },
            "ORACLE_SIG": {
                "value": "<OracleSig>",
                "type": "Signature"
            }
        },
        "creator_npub": pubkey,
    };

    // Usamos JSON.stringify para formatar o JSON na textarea
    let contractJson = JSON.stringify(defaultContractData, null, 2);

        isPublishing = true;
        statusMessage = 'Preparing nostr event...';

        let contractData;
        try {
            // validating JSON
            contractData = JSON.parse(contractJson);
        } catch (e) {
            statusMessage = 'Error: Json contract is invalid, verify sintax.';
            isPublishing = false;
            return;
        }

        try {
            const event:any = new NDKEvent(ndk);
            event.kind = 1;
            event.content = contractJson; // JSON as a string

            event.tags = [
                ['t', 'mare-nostrum'], // Main tag for search
                ['t', 'simplicity-contract'],
                ['t', 'simplicity'],
                ['t', 'mare-nostrum-contract'],
                ['t', 'liquid']
            ];

            // Dynamic tags based on the json
            if (contractData.network) {
                event.tags.push(['t', contractData.network]);
            }
            if (contractData.contractAddress) {
                event.tags.push(['r', contractData.contractAddress]);
            }

            statusMessage = 'Waiting for NIP07 signing...';
            
            // Signing event with NIP07
            await event.sign();

            statusMessage = 'Publishing in relays...';

            await event.publish();

            statusMessage = `Sucesso! Contrato propagado. Event ID: ${event.id}`;
            publishedEventJson = JSON.stringify(event.rawEvent(), null, 2);
        } catch (e:any) {
            console.error(e);
            statusMessage = `Error when publishing: ${e.message}`;
        } finally {
            isPublishing = false;
            published = true;
        }
    }
</script>

<div class="max-w-6xl mx-auto h-screen my-auto px-8 py-8">
    <div class="w-full text-center mx-auto mt-12 mb-8">
        <div class="inline-flex gap-2 items-center">
            <img src="/logotransparent.png" alt="MareNostrum logo" class="h-32" />
            <h1 class="text-5xl font-semibold text-white">MareNostrum</h1>
        </div>
        <p class="mt-2 text-gray-300">
            A Nostr propagation, exploration and fund layer for Liquid Simplicity contracts.
        </p>
    </div>

    {#if pubkey && published}
        <div class="max-w-6xl mx-auto p-6 bg-white rounded-xl shadow-xl border border-gray-200">
            <div class="text-center">
                <svg class="mx-auto h-12 w-12 text-green-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <h1 class="mt-2 text-2xl font-bold tracking-tight text-gray-900">
                    Simplicty Contract was Propagated on Nostr!
                </h1>
                <p class="mt-2 text-md text-gray-600">
                    The event with your contract was propagated on Nostr, now anyone can find and fund and/or spend your Simplicity contract.
                </p>
            </div>

            <div class="my-6 border-t border-gray-200"></div>

            <div>
                <label for="event-json" class="block text-sm font-medium text-gray-700">
                    Event JSON:
                </label>
                
                <pre class="
                    mt-1 
                    block 
                    w-full 
                    rounded-md 
                    bg-gray-900 
                    p-4 
                    text-sm 
                    text-green-400 
                    shadow-inner 
                    overflow-x-auto
                ">
                    <code>{publishedEventJson}</code>
                </pre>
            </div>
        </div>
    {/if}

    {#if pubkey && !published}
        <div class="p-4 space-y-8">
            <div class="border rounded-lg bg-white text-gray-700 p-6 space-y-6">
                {#if userProfile?.name}
                    <div class="md:flex gap-4 items-center">
                        <img src={userProfile.image} alt={userProfile.name} class="rounded-full h-24">
                        <div class="space-y-1">
                            <p class="font-semibold text-xl">{userProfile.name}</p>
                            <p>{userProfile.about}</p>
                        </div>
                    </div>
                    <p class="text-xs text-gray-600"><strong>Npub:</strong> {pubkey}</p>
                {/if}
            </div>
            {#if contractType && !addressReturned}
                <div class="mx-auto p-6 space-y-4 border rounded-lg bg-white text-gray-800">
                    <h1 class="text-xl font-bold mb-2">Submit Block Data</h1>

                    <div class="flex flex-col space-y-2">
                        <label for="blockHeight" class="font-medium">Contract name</label>
                        <input
                            id="name"
                            type="string"
                            bind:value={name}
                            placeholder="My lovely contract..."
                            class="border p-2 rounded-lg bg-white"
                        />
                    </div>

                    <div class="flex flex-col space-y-2">
                        <label for="blockHeight" class="font-medium">Brief description</label>
                        <input
                            id="description"
                            type="string"
                            bind:value={description}
                            placeholder="A contract for..."
                            class="border p-2 rounded-lg bg-white"
                        />
                    </div>

                    <div class="flex flex-col space-y-2">
                        <label for="blockHeight" class="font-medium">Block Height</label>
                        <input
                            id="blockHeight"
                            type="number"
                            bind:value={blockHeight}
                            placeholder="e.g. 853420"
                            class="border p-2 rounded-lg bg-white"
                        />
                    </div>

                    <div class="flex flex-col space-y-2">
                        <label for="price" class="font-medium">Price (USD)</label>
                        <input
                            id="price"
                            type="number"
                            step="0.01"
                            bind:value={price}
                            placeholder="e.g. 69000.50"
                            class="border p-2 rounded-lg bg-white"
                        />
                    </div>

                    <button
                        class="bg-blue-600 text-white p-2 rounded-lg w-full hover:bg-blue-700"
                        on:click={submitForm}
                    >
                        Submit
                    </button>
                </div>
            {:else if contractType && addressReturned &&!txid}
                <div class="mx-auto p-6 space-y-4 border rounded-lg bg-white text-gray-800">
                    <p>{addressReturned.message}</p>
                    <h1><strong>Address:</strong> {addressReturned?.address}</h1>
                    <h2><strong>Hex:</strong> {addressReturned?.program_hex}</h2>
                    <h3><strong>Contract name:</strong> {addressReturned?.name}</h3>
                    <h4><strong>Contract description:</strong> {addressReturned.description}</h4>
                    <h5>Target height: {addressReturned?.blockHeight}</h5>
                    <h6>Target price: {addressReturned?.price}</h6>
                    <button 
                        on:click={fundContract}
                        class="bg-green-500 text-white py-2 px-3 w-full rounded-md hover:bg-green-600"
                    >Fund contract with faucet coins</button>
                </div>
            {:else if contractType && addressReturned && txid}
                <div>
                    <div class="mb-6">
                        <h1 class="text-3xl font-semibold">Address funded!</h1>
                        <p><strong>txid:</strong> {txid}</p>
                    </div>
                    <button
                        on:click={publishContract}
                        class="bg-blue-600 text-white p-2 rounded-lg w-full hover:bg-blue-700"
                    >Propagate contract to Nostr</button>
                </div>
            {:else}
                <div class="mx-auto p-6 space-y-4 border rounded-lg bg-white">
                    <p class="text-lg font-semibold text-gray-700">Choose the contract type you're going to propagate:</p>
                    <button
                        on:click={() => contractType = !contractType}
                        class="bg-green-500 text-white py-2 px-3 w-full rounded-md hover:bg-green-600"
                    >
                        Prize Vault
                    </button>
                    <p class="text-gray-700 text-sm" mb-6>This contract is created to be spent by anyone as long as the Bitcoin price is above the defined target and the blockheight also above a determined target.</p>
                    <span class="text-xs text-gray-500 mt-4">(Sorry, we only got Hodl Vault contracts at the moment 😅)</span>
                </div>
            {/if}
        </div>
    {:else if !published}
        <div class="border rounded-xl p-6 max-w-lg mx-auto bg-white text-gray-700">
            <p class="my-4 text-center">Login with your Nostr account using the NIP07 browser extension.</p>
            <button
            class="bg-green-500 text-white py-2 px-3 w-full rounded-md hover:bg-green-600"
            on:click={connect}
            >
                Login with Nostr
            </button>
        </div>
    {/if}

    {#if error}
        <p class="text-red-500 mt-2">{error}</p>
    {/if}

</div>