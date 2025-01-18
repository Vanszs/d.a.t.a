import { Plugin } from "@elizaos/core";
import { ethereumDataProvider } from "./providers/ethereum/database";

export const onchainDataPlugin: Plugin = {
    name: "onchain data plugin",
    description: "Enables onchain data fetching",
    actions: [],
    providers: [ethereumDataProvider],
    evaluators: [],
    services: [],
    clients: [],
};
