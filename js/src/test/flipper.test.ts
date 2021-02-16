import * as assert from "assert";
import { burrow } from "../test";
import fs from 'fs';

describe('Wasm flipper:', function () {
    let TestContract

    before(async () => {
        const abi: any[] = JSON.parse(fs.readFileSync('src/test/flipper.abi', 'utf-8'))
        const wasm: string = fs.readFileSync('src/test/flipper.wasm').toString('hex')
        TestContract = await burrow.contracts.deploy(abi, wasm, null, true)
    })

    it('Flip', async () => {
        let output = await TestContract.get()
        assert.strictEqual(output[0], true)
        await TestContract.flip()
        output = await TestContract.get()
        assert.strictEqual(output[0], false)
    })
})
