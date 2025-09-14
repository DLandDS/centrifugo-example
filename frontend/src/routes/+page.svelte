<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Centrifuge } from 'centrifuge';

	interface Message {
		id: string;
		topic: string;
		content: string;
		author: string;
		timestamp: string;
	}

	let centrifuge: Centrifuge | null = null;
	let connected = false;
	let currentTopic = 'general';
	let messages: Message[] = [];
	let newMessage = '';
	let username = '';
	let subscription: any = null;
	let subscriptions: { [key: string]: any } = {};
	let availableTopics: string[] = [];

	const API_BASE = 'http://localhost:8080';
	const CENTRIFUGO_URL = 'ws://localhost:8000/connection/websocket';

	onMount(async () => {
		// Load available topics
		try {
			const response = await fetch(`${API_BASE}/api/topics`);
			const data = await response.json();
			availableTopics = data.topics;
		} catch (error) {
			console.error('Failed to load topics:', error);
			availableTopics = ['general', 'tech', 'random'];
		}

		// Set default username
		username = `User${Math.floor(Math.random() * 1000)}`;
		
		initializeCentrifugo();
	});

	// Only reconnect when username changes after initial mount
	let previousUsername = '';
	$: if (username && username !== previousUsername && previousUsername !== '') {
		previousUsername = username;
		reconnectWithNewToken();
	} else if (previousUsername === '') {
		previousUsername = username;
	}

	async function reconnectWithNewToken() {
		if (centrifuge) {
			centrifuge.disconnect();
		}
		// Small delay to ensure disconnection is complete
		setTimeout(() => {
			initializeCentrifugo();
		}, 500);
	}

	onDestroy(() => {
		// Unsubscribe from all subscriptions
		Object.values(subscriptions).forEach(sub => {
			if (sub) {
				sub.unsubscribe();
			}
		});
		if (centrifuge) {
			centrifuge.disconnect();
		}
	});

	function initializeCentrifugo() {
		// Get JWT token first, then connect
		getTokenAndConnect();
	}

	async function getTokenAndConnect() {
		try {
			// Get JWT token from backend
			const response = await fetch(`${API_BASE}/api/token`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					user: username,
				}),
			});

			if (!response.ok) {
				throw new Error('Failed to get token');
			}

			const data = await response.json();
			const token = data.token;

			// Initialize Centrifuge with token
			centrifuge = new Centrifuge(CENTRIFUGO_URL, {
				token: token
			});

			centrifuge.on('connected', () => {
				connected = true;
				console.log('Connected to Centrifugo');
				joinTopic(currentTopic);
			});

			centrifuge.on('disconnected', () => {
				connected = false;
				console.log('Disconnected from Centrifugo');
			});

			centrifuge.on('error', (error) => {
				console.error('Centrifugo error:', error);
			});

			centrifuge.connect();
		} catch (error) {
			console.error('Failed to get token or connect:', error);
			// Fallback to anonymous connection if token fails
			console.log('Falling back to anonymous connection...');
			centrifuge = new Centrifuge(CENTRIFUGO_URL);

			centrifuge.on('connected', () => {
				connected = true;
				console.log('Connected to Centrifugo (anonymous)');
				joinTopic(currentTopic);
			});

			centrifuge.on('disconnected', () => {
				connected = false;
				console.log('Disconnected from Centrifugo');
			});

			centrifuge.on('error', (error) => {
				console.error('Centrifugo error:', error);
			});

			centrifuge.connect();
		}
	}

	function joinTopic(topic: string) {
		// Unsubscribe from current subscription
		if (subscription) {
			subscription.unsubscribe();
		}

		messages = [];
		currentTopic = topic;

		if (centrifuge && connected) {
			const channelName = `topic:${topic}`;
			
			// Reuse existing subscription if available
			if (subscriptions[channelName]) {
				subscription = subscriptions[channelName];
			} else {
				// Create new subscription only if it doesn't exist
				subscription = centrifuge.newSubscription(channelName);
				subscriptions[channelName] = subscription;

				subscription.on('publication', (ctx: any) => {
					// Only add message if this is the current topic
					if (currentTopic === topic) {
						const message: Message = ctx.data;
						messages = [...messages, message];
						scrollToBottom();
					}
				});

				subscription.on('error', (error: any) => {
					console.error('Subscription error:', error);
				});
			}

			// Subscribe to the channel
			subscription.subscribe();
		}
	}

	async function sendMessage() {
		if (!newMessage.trim() || !username.trim() || !connected || !subscription) return;

		try {
			// Create message object
			const message: Message = {
				id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
				topic: currentTopic,
				content: newMessage.trim(),
				author: username.trim(),
				timestamp: new Date().toISOString()
			};

			// Publish directly via WebSocket using Centrifuge client
			await subscription.publish(message);
			
			// Clear input after successful send
			newMessage = '';
		} catch (error) {
			console.error('Failed to send message via WebSocket:', error);
			alert('Failed to send message. Please check your connection.');
		}
	}

	function scrollToBottom() {
		setTimeout(() => {
			const messagesContainer = document.getElementById('messages-container');
			if (messagesContainer) {
				messagesContainer.scrollTop = messagesContainer.scrollHeight;
			}
		}, 100);
	}

	function formatTime(timestamp: string) {
		return new Date(timestamp).toLocaleTimeString();
	}

	function handleKeyPress(event: KeyboardEvent) {
		if (event.key === 'Enter' && !event.shiftKey) {
			event.preventDefault();
			sendMessage();
		}
	}
</script>

<main class="container">
	<h1>Centrifugo Messaging App</h1>
	
	<div class="status-bar">
		<div class="connection-status" class:connected class:disconnected={!connected}>
			{connected ? '🟢 Connected' : '🔴 Disconnected'}
		</div>
		<div class="user-info">
			<label>
				Username:
				<input bind:value={username} placeholder="Enter your name" />
			</label>
		</div>
	</div>

	<div class="topics-bar">
		<span>Topics:</span>
		{#each availableTopics as topic}
			<button 
				class="topic-btn" 
				class:active={currentTopic === topic}
				on:click={() => joinTopic(topic)}
			>
				#{topic}
			</button>
		{/each}
	</div>

	<div class="chat-container">
		<div class="messages" id="messages-container">
			{#each messages as message}
				<div class="message" class:own-message={message.author === username}>
					<div class="message-header">
						<span class="author">{message.author}</span>
						<span class="time">{formatTime(message.timestamp)}</span>
					</div>
					<div class="message-content">{message.content}</div>
				</div>
			{:else}
				<div class="no-messages">
					No messages yet. Be the first to send a message to #{currentTopic}!
				</div>
			{/each}
		</div>

		<div class="message-input">
			<input
				bind:value={newMessage}
				on:keydown={handleKeyPress}
				placeholder="Type your message here..."
				disabled={!connected || !subscription}
			/>
			<button on:click={sendMessage} disabled={!connected || !newMessage.trim() || !subscription}>
				Send
			</button>
		</div>
	</div>
</main>

<style>
	.container {
		max-width: 800px;
		margin: 0 auto;
		padding: 20px;
		font-family: Arial, sans-serif;
	}

	h1 {
		text-align: center;
		color: #333;
		margin-bottom: 20px;
	}

	.status-bar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 10px;
		background: #f5f5f5;
		border-radius: 8px;
		margin-bottom: 20px;
	}

	.connection-status {
		font-weight: bold;
	}

	.connection-status.connected {
		color: green;
	}

	.connection-status.disconnected {
		color: red;
	}

	.user-info label {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.user-info input {
		padding: 5px;
		border: 1px solid #ddd;
		border-radius: 4px;
	}

	.topics-bar {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-bottom: 20px;
		padding: 10px;
		background: #e3f2fd;
		border-radius: 8px;
	}

	.topic-btn {
		padding: 5px 12px;
		border: 1px solid #2196f3;
		background: white;
		color: #2196f3;
		border-radius: 16px;
		cursor: pointer;
		transition: all 0.2s;
	}

	.topic-btn:hover {
		background: #e3f2fd;
	}

	.topic-btn.active {
		background: #2196f3;
		color: white;
	}

	.chat-container {
		border: 1px solid #ddd;
		border-radius: 8px;
		height: 500px;
		display: flex;
		flex-direction: column;
	}

	.messages {
		flex: 1;
		overflow-y: auto;
		padding: 20px;
		background: #fafafa;
	}

	.message {
		margin-bottom: 15px;
		padding: 10px;
		background: white;
		border-radius: 8px;
		box-shadow: 0 2px 4px rgba(0,0,0,0.1);
	}

	.message.own-message {
		background: #e3f2fd;
		margin-left: 50px;
	}

	.message-header {
		display: flex;
		justify-content: space-between;
		margin-bottom: 5px;
		font-size: 0.9em;
	}

	.author {
		font-weight: bold;
		color: #2196f3;
	}

	.time {
		color: #666;
	}

	.message-content {
		word-wrap: break-word;
	}

	.no-messages {
		text-align: center;
		color: #666;
		font-style: italic;
		margin-top: 50px;
	}

	.message-input {
		display: flex;
		padding: 15px;
		border-top: 1px solid #ddd;
		background: white;
	}

	.message-input input {
		flex: 1;
		padding: 10px;
		border: 1px solid #ddd;
		border-radius: 4px;
		margin-right: 10px;
	}

	.message-input button {
		padding: 10px 20px;
		background: #2196f3;
		color: white;
		border: none;
		border-radius: 4px;
		cursor: pointer;
		font-weight: bold;
	}

	.message-input button:disabled {
		background: #ccc;
		cursor: not-allowed;
	}

	.message-input button:not(:disabled):hover {
		background: #1976d2;
	}
</style>
