const path=require('node:path');
module.exports=async({page,out,assert})=>{
 const original=await page.evaluate(()=>api('/settings'));
 const existing=await page.evaluate(async()=> (await api('/clients')).clients.map(c=>({id:c.id,zone:c.zone,public_key:c.public_key})));
 const created=[];
 try{
  await page.evaluate(()=>location.hash='settings');
  const input=page.locator('#account-label-length');
  await input.waitFor({state:'visible'});
  assert.equal(await input.inputValue(),'3');
  assert.equal(await input.getAttribute('min'),'3');
  assert.equal(await input.getAttribute('max'),'26');
  for(const invalid of ['0','2','27','3.5','']){
   await input.fill(invalid);
   assert.equal(await input.evaluate(el=>el.checkValidity()),false,'invalid label length accepted: '+invalid);
  }
  const save=async length=>{
   await input.fill(String(length));
   const response=page.waitForResponse(r=>r.url().endsWith('/api/settings')&&r.request().method()==='PUT');
   await page.locator('#settings-form [type=submit]').click();
   assert.equal((await response).status(),200);
  };
  await save(6);
  await page.reload();await page.locator('#app').waitFor({state:'visible'});
  await page.waitForFunction(()=>document.getElementById('account-label-length').value==='6');
  const first=await page.evaluate(()=>api('/clients','POST',{name:'Label length fixture',enabled:true,engines:'all'}));
  created.push(first.id);
  assert.equal(first.zone.split('.')[0].length,6);
  assert.equal(first.id.length,32);
  await save(3);
  const preserved=await page.evaluate(id=>api('/clients/'+id),first.id);
  assert.equal(preserved.zone,first.zone);
  assert.equal(preserved.public_key,first.public_key);
  const second=await page.evaluate(()=>api('/clients','POST',{name:'Three character fixture',enabled:true,engines:'all'}));
  created.push(second.id);
  assert.equal(second.zone.split('.')[0].length,3);
  await page.evaluate(id=>api('/clients/'+id+'/rotate','POST'),first.id);
  const rotated=await page.evaluate(id=>api('/clients/'+id),first.id);
  assert.equal(rotated.zone.split('.')[0].length,3);
  assert.equal(rotated.id,first.id);
  assert.notEqual(rotated.public_key,first.public_key);
  const after=await page.evaluate(async()=> (await api('/clients')).clients);
  for(const old of existing){
   const current=after.find(c=>c.id===old.id);
   assert.equal(current.zone,old.zone);assert.equal(current.public_key,old.public_key);
  }
  await page.emulateMedia({reducedMotion:'reduce'});
  for(const lang of ['fa','en']){
   if(lang==='en'){
    await page.locator('#app .language-toggle').click();
    await page.locator('html[lang=en]').waitFor();await page.locator('#app').waitFor({state:'visible'});
    await page.waitForFunction(()=>document.getElementById('account-label-length').value==='3');
   }
   assert.equal(await page.getByLabel(lang==='fa'?'طول شناسهٔ دامنهٔ حساب':'Account domain label length',{exact:true}).count(),1);
   for(const [width,height] of [[1440,1100],[375,1100],[812,375]]){
    await page.setViewportSize({width,height});await input.scrollIntoViewIfNeeded();
    assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1),'label settings overflow');
    await page.screenshot({path:path.join(out,`account-labels-${lang}-${width}.png`)});
   }
  }
 }finally{
  await page.evaluate(async({settings,ids})=>{for(const id of ids)await api('/clients/'+id,'DELETE');await api('/settings','PUT',settings);localStorage.setItem('whitedns-panel-language','fa');},{settings:original,ids:created});
  await page.emulateMedia({reducedMotion:'no-preference'});await page.setViewportSize({width:1440,height:1100});
  await page.reload();await page.locator('#app').waitFor({state:'visible'});
  await page.evaluate(()=>location.hash='clients');await page.locator('#page-clients').waitFor({state:'visible'});
  await page.evaluate(()=>refresh());
 }
 console.log('PASS: account-label setting, validation, persistence, 3/6-character creation, explicit rotation, preserved accounts, Persian/English and responsive bounds.');
};
