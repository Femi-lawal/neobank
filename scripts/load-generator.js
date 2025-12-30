#!/usr/bin/env node
/**
 * NeoBank Load Generator
 * 
 * Generates traffic to all backend services to populate
 * Prometheus metrics, Grafana dashboards, and OpenTelemetry traces.
 * 
 * Usage: node load-generator.js
 */

const BASE_URL = 'http://localhost';
const SERVICES = {
    identity: `${BASE_URL}:8081`,
    ledger: `${BASE_URL}:8082`,
    payment: `${BASE_URL}:8083`,
    product: `${BASE_URL}:8084`,
    card: `${BASE_URL}:8085`,
};

const DELAY_MS = parseInt(process.env.DELAY_MS || '0', 10);
const sleep = (ms) => new Promise(r => setTimeout(r, ms));

let authToken = '';
let userId = '';
let accountIds = [];

// Helper to make requests
async function request(url, options = {}) {
    const headers = {
        'Content-Type': 'application/json',
        ...(authToken ? { 'Authorization': `Bearer ${authToken}` } : {}),
        ...options.headers,
    };

    try {
        const response = await fetch(url, { ...options, headers });
        const data = await response.json().catch(() => ({}));
        return { status: response.status, data, ok: response.ok };
    } catch (error) {
        return { status: 0, error: error.message, ok: false };
    }
}

// Register and login with new user
async function registerAndLogin() {
    const timestamp = Date.now();
    const email = `loadgen-${timestamp}@neobank.com`;
    const password = 'password123';

    console.log(`📝 Registering user ${email}...`);
    await request(`${SERVICES.identity}/auth/register`, {
        method: 'POST',
        body: JSON.stringify({
            email,
            password,
            first_name: 'Load',
            last_name: 'Generator'
        }),
    });

    console.log('🔐 Logging in...');
    const res = await request(`${SERVICES.identity}/auth/login`, {
        method: 'POST',
        body: JSON.stringify({ email, password }),
    });

    if (res.ok && res.data.token) {
        authToken = res.data.token;
        userId = res.data.user?.id || 'load-gen-user';
        console.log('✅ Login successful');
        return true;
    }
    console.log('❌ Login failed:', res.data);
    return false;
}

// Generate identity service traffic
async function generateIdentityTraffic(iterations) {
    console.log(`\n📛 Identity Service - ${iterations} iterations`);
    for (let i = 0; i < iterations; i++) {
        // Health check
        await request(`${SERVICES.identity}/health`);

        // Get profile (corrected endpoint)
        await request(`${SERVICES.identity}/api/v1/me`);

        // Metrics endpoint
        await request(`${SERVICES.identity}/metrics`);

        if (DELAY_MS > 0) await sleep(DELAY_MS);
        process.stdout.write(`\r  Progress: ${i + 1}/${iterations}`);
    }
    console.log(' ✅');
}

// Generate ledger service traffic
async function generateLedgerTraffic(iterations) {
    console.log(`\n💰 Ledger Service - ${iterations} iterations`);
    for (let i = 0; i < iterations; i++) {
        // Get accounts
        const res = await request(`${SERVICES.ledger}/api/v1/accounts`);
        if (res.ok && res.data.length > 0) {
            accountIds = res.data.map(a => a.id);
        }

        // Get specific account (if we have one)
        if (accountIds.length > 0) {
            await request(`${SERVICES.ledger}/api/v1/accounts/${accountIds[0]}`);
        }

        // Health check
        await request(`${SERVICES.ledger}/health`);

        // Metrics
        await request(`${SERVICES.ledger}/metrics`);

        if (DELAY_MS > 0) await sleep(DELAY_MS);
        process.stdout.write(`\r  Progress: ${i + 1}/${iterations}`);
    }
    console.log(' ✅');
}

// Generate payment service traffic
async function generatePaymentTraffic(iterations) {
    console.log(`\n💸 Payment Service - ${iterations} iterations`);
    for (let i = 0; i < iterations; i++) {
        // Health check
        await request(`${SERVICES.payment}/health`);

        // Get transfers list
        await request(`${SERVICES.payment}/api/v1/transfer`);

        // Create a simulated transfer
        if (accountIds.length >= 2) {
            const from = accountIds[Math.floor(Math.random() * accountIds.length)];
            let to = accountIds[Math.floor(Math.random() * accountIds.length)];
            // Ensure different accounts
            while (to === from) {
                to = accountIds[Math.floor(Math.random() * accountIds.length)];
            }

            await request(`${SERVICES.payment}/api/v1/transfer`, {
                method: 'POST',
                body: JSON.stringify({
                    from_account_id: from,
                    to_account_id: to,
                    amount: (Math.random() * 100 + 1).toFixed(2),
                    currency: 'USD',
                    description: `Load Test Transfer ${Date.now()}`
                })
            });
        }

        // Metrics
        await request(`${SERVICES.payment}/metrics`);

        if (DELAY_MS > 0) await sleep(DELAY_MS);
        // process.stdout.write is messy in parallel, maybe log less frequently?
        if (i % 10 === 0) process.stdout.write(`\r  Payment Progress: ${i + 1}/${iterations}`);
    }
    console.log('\n  ✅ Payment Service Done');
}

// Generate product service traffic
async function generateProductTraffic(iterations) {
    console.log(`\n🏦 Product Service - ${iterations} iterations`);
    for (let i = 0; i < iterations; i++) {
        // Get all products
        await request(`${SERVICES.product}/api/v1/products`);

        // Health check
        await request(`${SERVICES.product}/health`);

        // Metrics
        await request(`${SERVICES.product}/metrics`);

        if (DELAY_MS > 0) await sleep(DELAY_MS);
        process.stdout.write(`\r  Progress: ${i + 1}/${iterations}`);
    }
    console.log(' ✅');
}

// Generate card service traffic
async function generateCardTraffic(iterations) {
    console.log(`\n💳 Card Service - ${iterations} iterations`);
    for (let i = 0; i < iterations; i++) {
        // Get cards
        await request(`${SERVICES.card}/api/v1/cards`);

        // Health check
        await request(`${SERVICES.card}/health`);

        // Metrics
        await request(`${SERVICES.card}/metrics`);

        if (DELAY_MS > 0) await sleep(DELAY_MS);
        process.stdout.write(`\r  Progress: ${i + 1}/${iterations}`);
    }
    console.log(' ✅');
}

// Create some accounts
async function createTestAccounts(count) {
    console.log(`\n📊 Creating ${count} test accounts...`);
    const accountTypes = ['checking', 'savings', 'investment'];
    const currencies = ['USD', 'EUR', 'GBP'];

    for (let i = 0; i < count; i++) {
        await request(`${SERVICES.ledger}/api/v1/accounts`, {
            method: 'POST',
            body: JSON.stringify({
                name: `Test Account ${Date.now()}-${i}`,
                type: accountTypes[i % accountTypes.length],
                currency_code: currencies[i % currencies.length],
            }),
        });
        process.stdout.write(`\r  Created: ${i + 1}/${count}`);
    }
    console.log(' ✅');
}

// Main execution
async function main() {
    console.log('═══════════════════════════════════════════════════════════');
    console.log('  🏦 NeoBank Load Generator');
    console.log('  Generating traffic for Prometheus/Grafana/OpenTelemetry');
    console.log('═══════════════════════════════════════════════════════════');

    // Register & Login first
    const loggedIn = await registerAndLogin();
    if (!loggedIn) {
        console.log('\n⚠️  Could not login. Some requests may fail with 401.');
    }

    // Configuration
    const iterations = parseInt(process.env.ITERATIONS || '50', 10);
    const rounds = parseInt(process.env.ROUNDS || '3', 10);

    console.log(`\n📌 Configuration: ${rounds} rounds × ${iterations} iterations per service`);
    console.log(`   Total requests: ~${rounds * iterations * 5 * 4} requests\n`);

    for (let round = 1; round <= rounds; round++) {
        console.log(`\n══════════════════════════════════════════════════════════`);
        console.log(`  Round ${round}/${rounds}`);
        console.log(`══════════════════════════════════════════════════════════`);

        // Create some accounts each round
        await createTestAccounts(5);

        // Fetch accounts to populate accountIds for transfers
        const accRes = await request(`${SERVICES.ledger}/api/v1/accounts`);
        if (accRes.ok && accRes.data.length > 0) {
            accountIds = accRes.data.map(a => a.id);
        }

        console.log(`\n🚀 Launching traffic generators in parallel...`);
        // Generate traffic to all services in parallel
        await Promise.all([
            generateIdentityTraffic(iterations),
            generateLedgerTraffic(iterations),
            generatePaymentTraffic(iterations),
            generateProductTraffic(iterations),
            generateCardTraffic(iterations)
        ]);

        // Small delay between rounds
        if (round < rounds) {
            console.log('\n⏳ Waiting 2 seconds before next round...');
            await new Promise(resolve => setTimeout(resolve, 2000));
        }
    }

    console.log('\n═══════════════════════════════════════════════════════════');
    console.log('  ✅ Load generation complete!');
    console.log('');
    console.log('  📊 View dashboards:');
    console.log('     • Prometheus: http://localhost:9090');
    console.log('     • Grafana:    http://localhost:3000 (admin/admin)');
    console.log('     • Jaeger:     http://localhost:16686');
    console.log('═══════════════════════════════════════════════════════════\n');
}

main().catch(console.error);
