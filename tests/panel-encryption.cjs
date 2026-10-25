module.exports=async({page,assert})=>{
 await page.evaluate(()=>location.hash='clients');
 await page.locator('#page-clients .new-client').click();
 const select=page.locator('#client-encryption-method'),trigger=page.locator('#client-encryption-method-trigger');
 assert.equal(await select.inputValue(),'5');
 await page.locator('#client-form [name=name]').fill('Encryption fixture');
 await trigger.click();await page.locator('#panel-select-options [data-index="1"]').click();
 assert.equal(await select.inputValue(),'1');
 const save=async()=>{const response=page.waitForResponse(r=>r.url().includes('/api/clients')&&['PUT','POST'].includes(r.request().method()));await page.locator('#client-form [type=submit]').click();const result=await response;assert.equal(result.status(),result.request().method()==='POST'?201:200);await page.locator('#client-dialog').waitFor({state:'hidden'});};
 await save();
 let c=await page.evaluate(async()=> (await api('/clients')).clients.find(c=>c.name==='Encryption fixture'));
 try{
  assert.equal(c.encryption_method,1);
  const profiles=await page.evaluate(id=>api('/clients/'+id+'/link'),c.id);
  for(const engine of ['cotten','stormdns','masterdns']){
   const body=JSON.parse(Buffer.from(profiles.links[engine].split('://')[1],'base64url'));
   assert.equal(body.profile.server.encryption_method,1);
  }
  await page.reload();await page.locator('#app').waitFor({state:'visible'});
  await page.evaluate(()=>location.hash='clients');
  await page.locator('#clients-body tr').filter({hasText:'Encryption fixture'}).getByRole('button',{name:'ویرایش',exact:true}).click();
  assert.equal(await select.inputValue(),'1');
  await trigger.click();await page.locator('#panel-select-options [data-index="0"]').click();
  await save();
  c=await page.evaluate(id=>api('/clients/'+id),c.id);assert.equal(c.encryption_method,5);
 }finally{await page.evaluate(id=>api('/clients/'+id,'DELETE'),c.id);await page.evaluate(()=>refresh());}
 console.log('PASS: encryption default, XOR selection, exported profiles, persistence and AES switch.');
};
