package main

import (
	"smppizdez/account"
	"smppizdez/coding"
	"strconv"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

var bindTypes = []account.BindType{account.Transceiver, account.Receiver, account.Transmitter}

func getBindTypeIndex(t account.BindType) int {
	for i, typ := range bindTypes {
		if t == typ {
			return i
		}
	}
	return 0
}

func bindTypeToString(typ account.BindType) string {
	switch typ {
	case account.Transceiver:
		return "transceiver"
	case account.Receiver:
		return "receiver"
	case account.Transmitter:
		return "transmitter"
	default:
		return "unknown"
	}
}

var defaultCodings = []coding.Coding{coding.GSM7, coding.GSM8}

func getDefaultCodingIndex(c coding.Coding) int {
	for i, cod := range defaultCodings {
		if c == cod {
			return i
		}
	}
	return 0
}

type accountDialog struct {
	window           *gtk.ApplicationWindow
	label            *gtk.Label
	hostEntry        *gtk.Entry
	portEntry        *gtk.Entry
	tlsSwitch        *gtk.Switch
	systemIdEntry    *gtk.Entry
	passwordEntry    *gtk.Entry
	systemTypeEntry  *gtk.Entry
	bindTypeSelector *gtk.ComboBox
	codingSelector   *gtk.ComboBox
	callback         func(*account.Account)
}

type accountsContext struct {
	accounts []account.Account
	dialog   accountDialog
	tree     *gtk.TreeView
	repo     account.Repository
}

func loadAccounts(ctx *accountsContext) {
	var err error
	ctx.accounts, err = ctx.repo.GetAccounts()
	if err != nil {
		errorDialog("Failed to load accounts: %v", err)
		if len(ctx.accounts) == 0 {
			return
		}
	}

	store, _ := gtk.ListStoreNew(
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		glib.TYPE_INT,
		glib.TYPE_BOOLEAN,
		glib.TYPE_STRING,
	)
	for _, account := range ctx.accounts {
		iter := store.Append()
		accountToIter(store, iter, &account)
	}
	ctx.tree.SetModel(store)
	store.Unref()

	renderer, _ := gtk.CellRendererTextNew()
	columnNames := []string{"System ID", "Host", "Port", "TLS", "Bind Type"}
	for i, name := range columnNames {
		column, _ := gtk.TreeViewColumnNewWithAttribute(name, renderer, "text", i)
		ctx.tree.AppendColumn(column)
	}
}

func initAccountsMenu(ctx *accountsContext, builder *gtk.Builder) {
	addItem := getMenuItemById(builder, "add_account_item")
	addItem.Connect("button_release_event", func() { addAccountHandler(ctx) })
	editItem := getMenuItemById(builder, "edit_account_item")
	editItem.Connect("button_release_event", func() { editAccountHandler(ctx) })
	delItem := getMenuItemById(builder, "del_account_item")
	delItem.Connect("button_release_event", func() { deleteAccountHandler(ctx) })
}

func accountToIter(store *gtk.ListStore, iter *gtk.TreeIter, account *account.Account) {
	values := []any{
		account.SystemID,
		account.Host,
		int(account.Port),
		account.TLS,
		bindTypeToString(account.BindType),
	}
	store.Set(iter, []int{0, 1, 2, 3, 4}, values)
}

func addAccountHandler(ctx *accountsContext) {
	ctx.dialog.run(nil, func(account *account.Account) {
		err := ctx.repo.CreateAccount(account)
		if err != nil {
			errorDialog(err.Error())
		}

		modelI, _ := ctx.tree.GetModel()
		store := modelI.(*gtk.ListStore)
		iter := store.Append()
		accountToIter(store, iter, account)
		ctx.accounts = append(ctx.accounts, *account)
	})
}

func getSelectedAccountIdx(ctx *accountsContext) (int, *gtk.TreePath) {
	path, _ := ctx.tree.GetCursor()
	idx := path.GetIndices()[0]
	if idx < 0 || idx >= len(ctx.accounts) {
		errorDialog(
			"Current tree view index (%d) is greater than number of accounts (%d)",
			idx,
			len(ctx.accounts),
		)
		return 0, nil
	}
	return idx, path
}

func editAccountHandler(ctx *accountsContext) {
	idx, path := getSelectedAccountIdx(ctx)
	if path == nil {
		return
	}

	ctx.dialog.run(&ctx.accounts[idx], func(account *account.Account) {
		err := ctx.repo.UpdateAccount(account)
		if err != nil {
			errorDialog("Failed to update account: %v", err)
		}

		modelI, _ := ctx.tree.GetModel()
		store := modelI.(*gtk.ListStore)
		iter, err := store.GetIter(path)
		if err != nil {
			errorDialog("Failed to update account: %v", err)
			return
		}
		accountToIter(store, iter, account)
		ctx.accounts[idx] = *account
	})
}

func deleteAccountHandler(ctx *accountsContext) {
	idx, path := getSelectedAccountIdx(ctx)
	if path == nil {
		return
	}

	modelI, _ := ctx.tree.GetModel()
	store := modelI.(*gtk.ListStore)
	iter, err := store.GetIter(path)
	if err != nil {
		errorDialog("Failed to delete account: %v", err)
		return
	}
	store.Remove(iter)
	delId := ctx.accounts[idx].ID
	ctx.accounts = append(ctx.accounts[:idx], ctx.accounts[idx+1:]...)

	err = ctx.repo.DeleteAccount(delId)
	if err != nil {
		errorDialog("Failed to delete account: %v", err)
	}
}

func initAccountDialog(d *accountDialog, builder *gtk.Builder) {
	accountDialogI, _ := builder.GetObject("account_dialog")
	d.window = accountDialogI.(*gtk.ApplicationWindow)
	d.window.Connect("delete_event", func() bool {
		d.window.Hide()
		return true
	})

	accountDialogOKBtn := getButtonById(builder, "account_dialog_ok_button")
	accountDialogOKBtn.Connect("pressed", func() {
		if account := d.validate(); account != nil {
			d.window.Hide()
			d.callback(account)
			d.callback = nil
		}
	})

	accountDialogCancelBtn := getButtonById(builder, "account_dialog_cancel_button")
	accountDialogCancelBtn.Connect("pressed", func() {
		d.window.Hide()
		d.callback = nil
	})

	d.label = getLabelById(builder, "account_dialog_label")
	d.hostEntry = getEntryById(builder, "account_dialog_host_entry")
	d.portEntry = getEntryById(builder, "account_dialog_port_entry")
	d.tlsSwitch = getSwitchById(builder, "account_dialog_tls_switch")
	d.systemIdEntry = getEntryById(builder, "account_dialog_system_id_entry")
	d.passwordEntry = getEntryById(builder, "account_dialog_password_entry")
	d.systemTypeEntry = getEntryById(builder, "account_dialog_system_type_entry")
	d.bindTypeSelector = getComboById(builder, "account_dialog_bind_type_selector")
	d.codingSelector = getComboById(builder, "account_dialog_coding_selector")
}

func (d *accountDialog) resetStyles() {
	widgets := []*gtk.Widget{
		&d.hostEntry.Widget,
		&d.portEntry.Widget,
		&d.systemIdEntry.Widget,
		&d.passwordEntry.Widget,
		&d.systemTypeEntry.Widget,
		&d.bindTypeSelector.Widget,
		&d.codingSelector.Widget,
	}

	for _, widget := range widgets {
		widget.SetTooltipText("")
		ctx, _ := widget.GetStyleContext()
		ctx.RemoveClass("invalid-entry")
	}
}

func (d *accountDialog) reset() {
	d.resetStyles()
	d.hostEntry.SetText("")
	d.portEntry.SetText("2775")
	d.tlsSwitch.SetActive(false)
	d.systemIdEntry.SetText("")
	d.passwordEntry.SetText("")
	d.systemTypeEntry.SetText("")
	d.bindTypeSelector.SetActive(0)
	d.codingSelector.SetActive(0)
}

func (d *accountDialog) validate() *account.Account {
	d.resetStyles()

	isValid := true
	host, ok := checkEntryPresence(d.hostEntry, "Host")
	isValid = isValid && ok

	portU64, ok := checkEntryNumerical(d.portEntry, 16, "Port")
	isValid = isValid && ok
	port := uint16(portU64)

	systemId, ok := checkEntryPresence(d.systemIdEntry, "System ID")
	isValid = isValid && ok

	password, ok := checkEntryPresence(d.passwordEntry, "Password")
	isValid = isValid && ok

	systemType, _ := d.systemTypeEntry.GetText()
	tls := d.tlsSwitch.GetActive()

	bindTypeIdx := getComboIndex(d.bindTypeSelector)
	bindType := bindTypes[bindTypeIdx]

	defaultCodingIdx := getComboIndex(d.codingSelector)
	defaultCoding := defaultCodings[defaultCodingIdx]

	if isValid {
		account := &account.Account{
			Host:          host,
			Port:          port,
			TLS:           tls,
			SystemID:      systemId,
			Password:      password,
			SystemType:    systemType,
			BindType:      bindType,
			DefaultCoding: defaultCoding,
		}
		return account
	}
	return nil
}

func (d *accountDialog) run(acc *account.Account, callback func(*account.Account)) {
	d.reset()

	if acc != nil {
		d.label.SetText("Edit account")
		d.callback = func(account *account.Account) {
			account.ID = acc.ID
			callback(account)
		}

		d.hostEntry.SetText(acc.Host)
		d.portEntry.SetText(strconv.Itoa(int(acc.Port)))
		d.tlsSwitch.SetActive(acc.TLS)
		d.systemIdEntry.SetText(acc.SystemID)
		d.passwordEntry.SetText(acc.Password)
		d.systemTypeEntry.SetText(acc.SystemType)
		d.bindTypeSelector.SetActive(getBindTypeIndex(acc.BindType))
		d.codingSelector.SetActive(getDefaultCodingIndex(acc.DefaultCoding))
	} else {
		d.label.SetText("Add new account")
		d.callback = callback
	}

	d.window.Present()
}

func (ctx *accountsContext) selectAccountHandler(tree *gtk.TreeView, path *gtk.TreePath) {
	acc := &ctx.accounts[path.GetIndices()[0]]
	submitSmStartSessionCallback(acc)
}

func initAccountsList(builder *gtk.Builder, repo account.Repository) {
	accountsListI, _ := builder.GetObject("accounts_list")
	tree := accountsListI.(*gtk.TreeView)
	ctx := accountsContext{
		tree: tree,
		repo: repo,
	}
	initAccountDialog(&ctx.dialog, builder)

	accountsMenuI, _ := builder.GetObject("accounts_menu")
	menu := accountsMenuI.(*gtk.Menu)
	initAccountsMenu(&ctx, builder)

	tree.Connect("row_activated", ctx.selectAccountHandler)

	loadAccounts(&ctx)
	tree.Connect("button_press_event", func(_ *gtk.TreeView, event *gdk.Event) bool {
		btnEvent := gdk.EventButtonNewFromEvent(event)
		if btnEvent.Button() == gdk.BUTTON_SECONDARY {
			menu.PopupAtPointer(event)
			return true
		}
		return false
	})
}
