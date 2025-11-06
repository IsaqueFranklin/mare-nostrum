<script lang="ts">
	import NDK, { NDKNip07Signer, type NDKUserProfile } from "@nostr-dev-kit/ndk";
	import { browser } from '$app/environment';
	import { onMount } from "svelte";
    import axios from "axios"

    interface ContractResponse {
        message: string;
        blockHeight: number;
        price: number;
        program_hex: string;
        address: string;
    }
	let ndk: NDK | null = null;
	let signer: NDKNip07Signer | null = null;
	let userProfile: NDKUserProfile | null = null;
	let pubkey: string | null = null;
	let error: string | null = null;

    let contractType: boolean = false;
    let addressReturned: ContractResponse | null = null;
    let txid: string | null = null;

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

	async function submitForm() {
		if (blockHeight === null || price === null) {
			alert("Please fill in both fields.");
			return;
		}

		//const data = { blockHeight, price };

		try {
			const { data } = await axios.post("http://localhost:4000/generate-contract", {
				blockHeight,
				price
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

    {#if pubkey}
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
                    <h3>Target height: {addressReturned?.blockHeight}</h3>
                    <h4>Target price: {addressReturned?.price}</h4>

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
    {:else}
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
        <p class="text-red-500 mt-2">⚠️ {error}</p>
    {/if}

</div>