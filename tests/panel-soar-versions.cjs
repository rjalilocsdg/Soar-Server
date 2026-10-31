module.exports=async function({page,assert}){
 const row=page.locator('#clients-body tr').first();await row.getByRole('button',{name:'ویرایش',exact:true}).click();
 assert.equal(await page.locator('#client-soar-version').inputValue(),'0.2.4');
 await page.locator('#client-soar-version-trigger').click();await page.locator('#panel-select-options').getByRole('option',{name:'v0.2.3',exact:true}).click();
 await page.locator('#client-form button[type=submit]').click();await page.locator('#client-dialog').waitFor({state:'hidden'});
 await row.getByRole('button',{name:'ویرایش',exact:true}).click();assert.equal(await page.locator('#client-soar-version').inputValue(),'0.2.3');
 await page.locator('#client-soar-version-trigger').click();await page.locator('#panel-select-options').getByRole('option',{name:'v0.2.4 — پیش‌فرض',exact:true}).click();
 await page.locator('#client-form button[type=submit]').click();await page.locator('#client-dialog').waitFor({state:'hidden'});
 console.log('PASS: per-account Soar version selection persists and can return to default.');
};
