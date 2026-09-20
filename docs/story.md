# Why Hensu exists

You ask an assistant whether the database password is set. It does the obvious thing: it opens the `.env`. Now that password is in the chat, in the session log, and in a model's context. Three places it does not leave.

Nobody did anything wrong. The question was small — is it set or not? — and the answer came back whole.

Hensu did not start as a product. It came out of day-to-day work on **Naoki** at [Rosvelt](https://rosvelt.com/): a script inside the repo, `scripts/service/dotenv`, that kept growing because everything ended up needing the same thing — take the service's config, put it in the environment, run the program. Start the service. Seed the database. Run the tests. Open the workbench. Six different things doing the same dance, each carrying its own copy of the dance.

That script already knew how to hide values — it could print `***` instead of the secret. But hiding was optional. You had to pick the careful verb, and that works as long as the only reader is someone who knows what they are doing.

The readers changed. A model asked to check whether a key is configured is not being careless — it is doing the obvious thing, and opening the file *is* the obvious thing. Asking it to remember the prudent command is asking it to have a habit. A habit is not a mechanism.

So the default flipped. A value now comes back clipped, always, unless somebody types the word:

```bash
hensu get PORT            # {"PORT": {"value": "***", "defined": true}}
hensu -r trust get PORT   # {"PORT": {"value": "8080", "defined": true}}
```

And when the value is genuinely needed, it is almost never needed by whoever is asking — it is needed by a program. So give it to the program and be done. `hensu exec -- go run .` runs your app with the config inside it, and you saw nothing.

This is not a cage, and that is worth saying out loud. Anyone with a terminal can `cat .env` and see everything. No CLI can prevent that. What changed is something else: the comfortable path stopped being the leaky one, and a leak became something a person deliberately typed, in a command you can go looking for tomorrow.

The family's second published CLI: GoDo first, goxdi before all of it. godo did not stay decorative — it runs Hensu's own gate, `godo ci` green before every tag.

The name is **変数** — *hensū*, variable. 変 is change, 数 is value. That is what a config file is: values that change from machine to machine, named so a program can find them.

If you trust everyone in the room, you do not need any of this. I stopped being sure who is in the room.

[Getting started](./getting-started.md) · [Español](./story.es.md)
