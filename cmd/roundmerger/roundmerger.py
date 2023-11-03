import asyncpg
import sys
import asyncio
import os

if len(sys.argv) != 4:
    print("Usage: python3 roundmerger.py <database> <round A> <round B>")
    print("Moves the prices from round B to round A as a second index")
    exit(1)

async def main():
    database = sys.argv[1]
    roundA = sys.argv[2]
    roundB = sys.argv[3]
    pool = await asyncpg.create_pool(database=database)

    # Get game id from round A, round B
    gameA = await pool.fetchval("SELECT id FROM games WHERE id::text = $1 OR code = $1", roundA)
    gameB = await pool.fetchval("SELECT id FROM games WHERE id::text = $1 OR code = $1", roundB)

    if not gameA:
        print("gameA not found")
        exit(1)

    if not gameB:
        print("gameB not found")
        exit(1)

    # Get all prices of round B
    pricesB = await pool.fetch("SELECT id, ticker, prices FROM stocks WHERE game_id = $1", gameB)
    
    stock_pc = {}

    for price in pricesB:
        # Get round two prices for the stock
        pricesA = await pool.fetchval("SELECT prices FROM stocks WHERE game_id = $1 AND ticker = $2", gameA, price['ticker'])
        idA = await pool.fetchval("SELECT id FROM stocks WHERE game_id = $1 AND ticker = $2", gameA, price['ticker'])
        resultant = pricesA + [price['prices'][-1]]
        print(f"{price['id']} - {price['prices'][-1]} will be added to {pricesA} to give {resultant}")

        stock_pc[price['id']] = resultant

        await pool.execute("UPDATE stocks SET prices = $1 WHERE id = $2", resultant, idA)

asyncio.run(main())