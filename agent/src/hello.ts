/**
 * Day 0 exercise: your first LLM call, with STRUCTURED OUTPUT.
 *
 * Instead of asking the model for free text and hoping it's parseable, we give
 * it a "tool" whose input schema is exactly the JSON shape we want, then force
 * the model to call it. The tool input *is* our structured result.
 *
 * Run:  npm run hello
 *
 * Things to try once it works (this is the learning):
 *  1. Remove the category definitions from the system prompt. Does accuracy drop?
 *  2. Add a tricky example: "PAYPAL *STEAMGAMES". Where should it go?
 *  3. Change temperature, run it 5 times. Does the output change?
 *  4. Print response.usage. How many tokens did this cost?
 */
import Anthropic from "@anthropic-ai/sdk";
import "dotenv/config";

const client = new Anthropic(); // reads ANTHROPIC_API_KEY from .env

export const CATEGORIES = [
  "Income", "Housing", "Bills", "Subscriptions", "Groceries",
  "Eating Out", "Transport", "Shopping", "Other",
] as const;

const SYSTEM = `You categorise UK bank transactions for a personal finance app.

Category definitions:
- Income: salary, refunds, transfers in
- Housing: rent, mortgage
- Bills: energy, water, council tax, phone, broadband
- Subscriptions: recurring digital or membership services (streaming, gym, cloud storage)
- Groceries: supermarkets and convenience stores
- Eating Out: cafes, restaurants, pubs, takeaway and delivery apps
- Transport: public transport, fuel, taxis, parking
- Shopping: retail and online shopping not covered above
- Other: anything that doesn't fit

Also return a clean, human-readable merchant name (e.g. "TESCO STORES 3297" → "Tesco").`;

const categoriseTool: Anthropic.Tool = {
  name: "record_categories",
  description: "Record the merchant and category for each transaction, in the same order as given.",
  input_schema: {
    type: "object",
    properties: {
      results: {
        type: "array",
        items: {
          type: "object",
          properties: {
            description: { type: "string" },
            merchant: { type: "string" },
            category: { type: "string", enum: [...CATEGORIES] },
          },
          required: ["description", "merchant", "category"],
        },
      },
    },
    required: ["results"],
  },
};

const sample = [
  "TESCO STORES 3297 LONDON GB",
  "CARD PAYMENT TO PRET A MANGER",
  "NETFLIX.COM 866-579-7172",
  "TFL TRAVEL CH TFL.GOV.UK/CP",
  "AMZNMKTPLACE*2K4L91",
  "DD OCTOPUS ENERGY",
];

async function main() {
  const response = await client.messages.create({
    model: process.env.CARDO_MODEL ?? "claude-haiku-5-5",
    max_tokens: 1024,
    system: SYSTEM,
    tools: [categoriseTool],
    tool_choice: { type: "tool", name: "record_categories" }, // force structured output
    messages: [{ role: "user", content: sample.map((d, i) => `${i + 1}. ${d}`).join("\n") }],
  });

  const block = response.content.find((b) => b.type === "tool_use");
  if (!block || block.type !== "tool_use") throw new Error("Model did not call the tool");

  const { results } = block.input as { results: { description: string; merchant: string; category: string }[] };
  console.table(results);
  console.log("tokens:", response.usage);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
