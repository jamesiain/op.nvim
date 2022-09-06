<div align="center">

# op.nvim

<!-- panvimdoc-ignore-start -->

![Neovim version](https://img.shields.io/badge/Neovim-0.6-brightgreen?logo=neovim) ![1Password CLI V2](https://img.shields.io/badge/1Password%20CLI-V2-blue?logo=1password) [![GitHub license](https://img.shields.io/github/license/mrjones2014/op.nvim)](https://github.com/mrjones2014/op.nvim/blob/master/LICENSE)

[Prerequisites](#prerequisites) • [Install](#install) • [Configuration](#configuration) • [Commands](#commands) • [Features](#features) • [API](#api)

<!-- panvimdoc-ignore-end -->

</div>

1Password for Neovim! Create items using strings from the current buffer as fields,
and insert item reference URIs (e.g. `op://vault-name/item-name/field-name`)
directly from Neovim. Edit Secure Notes directly in Neovim. Works with biometric unlock!

<!-- panvimdoc-ignore-start -->

<details>
<summary>Screenshots and Gifs (click to expand)</summary>

**Secure Notes Editor**
![Secure Notes Editor](https://user-images.githubusercontent.com/8648891/188518923-1165ed52-4915-4443-9a8c-6285bf601055.gif)

**1Password Sidebar**

![1Password Sidebar](https://user-images.githubusercontent.com/8648891/188519356-ba166555-a587-4628-9756-520ca8fd8864.gif)

**Item Creation**

![Item creation](https://user-images.githubusercontent.com/8648891/188519718-8fbdde4b-14cc-4d2c-b534-83b7fa1829c9.gif)

</details>

<!-- panvimdoc-ignore-end -->

[More screenshots and demo gifs in the Wiki](https://github.com/mrjones2014/op.nvim/wiki/Screenshots-and-Gifs)!

## Prerequisites

**Required:**

- [1Password CLI v2](https://developer.1password.com/docs/cli/) installed

**Optional, but recommended:**

- [1Password 8 desktop app](https://1password.com/downloads/) (required to use biometric unlock for CLI)
- [Biometric unlock for CLI](https://developer.1password.com/docs/cli/get-started#turn-on-biometric-unlock) enabled (see [Using Token-Based Sessions](#using-token-based-sessions) if you do not use biometric unlock for CLI)
- A Neovim plugin to handle `vim.ui.select()` and `vim.ui.input()` &mdash; I recommend [telescope.nvim](https://github.com/nvim-telescope/telescope.nvim) paired with [dressing.nvim](https://github.com/stevearc/dressing.nvim)

### Windows Support

This plugin does not currently support Windows. I don't use Windows so I can't test on Windows.
However, I would happily accept Pull Requests adding Windows support, with a commitment to
ongoing maintenance from the PR author.

## Install

`packer.nvim`

```lua
use({ 'mrjones2014/op.nvim', run = 'make install' })
```

`vim-plug`

```VimL
Plug 'mrjones2014/op.nvim', { 'do': 'make install' }
```

No other setup is required if using biometric unlock for the 1Password CLI,
however there are a few settings you can change if needed. See [Configuration](#configuration).

## Configuration

Configuration can be set by calling `require('op').setup(config_table)`.

**The `require('op').setup()` function is idempotent** (i.e. can be called multiple times without side effects).

```lua
require('op').setup({
  -- you can change this to a full path if `op`
  -- is not on your $PATH
  op_cli_path = 'op',
  -- Whether to sign in on start.
  signin_on_start = false,
  -- show NerdFont icons in `vim.ui.select()` interfaces,
  -- set to false if you do not use a NerdFont or just
  -- don't want icons
  use_icons = true,
  -- command to use for opening URLs,
  -- can be a function or a string
  url_open_command = function()
    if vim.fn.has('mac') == 1 then
      return 'open'
    elseif vim.fn.has('unix') == 1 then
      return 'xdg-open'
    end
    return nil
  end,
  -- settings for op.nvim sidebar
  sidebar = {
    -- sections to include, available sections
    -- are 'favorites' and `secure_notes`
    'favorites',
    'secure_notes',
    -- sidebar width
    width = 40,
    -- put the sidebar on the right or left side
    side = 'right',
    -- keymappings for the sidebar buffer.
    -- can be a string mapping to a function from
    -- the module `op.sidebar.actions`,
    -- an editor command string, or a function.
    -- if you supply a function, a table with the following
    -- fields will be passed as an argument:
    -- {
    --   title: string,
    --   icon: string,
    --   type: 'header' | 'item'
    --   -- data will be nil if type == 'header'
    --   data: nil | {
    --       uuid: string,
    --       vault_uuid: string,
    --       category: string,
    --       url: string
    --     }
    -- }
    mappings = {
      -- if it's a Secure Note, open in op.nvim's Secure Notes editor;
      -- if it's an item with a URL, open & fill the item in default browser;
      -- otherwise, open in 1Password 8 desktop app
      ['<CR>'] = 'default_open',
      -- open in 1Password 8 desktop app
      ['go'] = 'open_in_desktop_app',
      -- edit in 1Password 8 desktop app
      ['ge'] = 'edit_in_desktop_app',
    },
  },
  -- Custom formatter function for statusline component
  statusline_fmt = function(account_name)
    if not account_name or #account_name == 0 then
      return ' 1Password: No active session'
    end

    return string.format(' 1Password: %s', account_name)
  end
  -- global_args accepts any arguments
  -- listed under "Global Flags" in
  -- `op --help` output.
  global_args = {
    -- use the item cache
    '--cache',
    -- print output with no color, since we
    -- aren't viewing the output directly anyway
    '--no-color',
  },
  -- Use biometric unlock by default,
  -- set this to false and also see
  -- "Using Token-Based Sessions" section
  -- of README.md if you don't use biometric
  -- unlock for CLI.
  biometric_unlock = true,
  -- settings for Secure Notes editor
  secure_notes = {
    -- prefix for buffer names when
    -- editing 1Password Secure Notes
    buf_name_prefix = '1P:',
  }
})
```

### Using Token-Based Sessions

If you do not use biometric unlock for the 1Password CLI, you can use token-based sessions.
**You must run `eval $(op signin)` _before_ launching Neovim** in order for `op.nvim` to be
able to access the session. You also **must** configure `op.nvim` with `biometric_unlock = false`.

## Commands

\* = Asynchronous \
† = Partially asynchronous

- `:OpSignin` \* - Choose a 1Password account to sign in with. Accepts account shorthand, signin address, account UUID, or user UUID as an optional argument.
- `:OpSignout` \* - End your current 1Password CLI session.
- `:OpWhoami` \* - Check which 1Password account your current CLI session is using.
- `:OpCreate` † - Create a new item using strings in the current buffer as fields.
- `:OpView` † - Open an item in the 1Password 8 desktop app.
- `:OpEdit` † - Open an item to the edit view in the 1Password 8 desktop app.
- `:OpOpen` - Select an item to open & fill in your default browser
- `:OpInsert` - Insert an item reference at current cursor position.
- `:OpNote` - Find and open a 1Password Secure Note item. Accepts `new` or `create` as an argument to create a new Secure Note.
- `:OpSidebar` \* - Toggle the 1Password sidebar open/closed. Accepts `refresh` as an argument to reload items.

### Lua API

All commands are also available as a Lua API as described below:

- `require('op').op_signin(account_identifier: string | nil)`
- `require('op).signout()`
- `require('op').op_whoami()`
- `require('op').op_create()`
- `require('op').op_view()`
- `require('op').op_edit()`
- `require('op').op_open()`
- `require('op').op_insert()`
- `require('op').op_note(create_new: boolean)`
- `require('op').op_sidebar(should_refresh: boolean)`

## Features

- Biometric unlock! Unlock 1Password with fingerprint or Apple watch from within Neovim
- Create items from strings in the current buffer
  - If the Treesitter query fails or there's no Treesitter parser for the current filetype, fallback to manual value input (if a Treesitter parser exists, please open an issue or PR so we can get the right query added!)
- Infer default field and item names based on field value patterns
- Open an item in the 1Password 8 desktop app
- Insert an item reference URI (e.g. `op://vault-name/item-name/field-name`)
- Switch between multiple 1Password accounts (only works with biometric unl
