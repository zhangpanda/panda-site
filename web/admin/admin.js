const API=location.origin+'/api';
let token=localStorage.getItem('ps_token')||'';
let sec='articles';
const $=id=>document.getElementById(id);
const esc=s=>{const d=document.createElement('div');d.textContent=s||'';return d.innerHTML};
function headers(){return{'Content-Type':'application/json','Authorization':'Bearer '+token}}
function toast(m){const t=$('toast');t.textContent=m;t.classList.remove('hidden');setTimeout(()=>t.classList.add('hidden'),2000)}

// === Auth ===
async function doLogin(){
  const u=$('loginUser').value,p=$('loginPwd').value;
  // 尝试初始化
  let res=await fetch(API+'/init',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({username:u,password:p})});
  if(res.status===201){const d=await res.json();token=d.token;localStorage.setItem('ps_token',token);init();return}
  // 正常登录
  res=await fetch(API+'/login',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({username:u,password:p})});
  const d=await res.json();
  if(d.token){token=d.token;localStorage.setItem('ps_token',token);init()}
  else{$('loginErr').textContent=d.error||'登录失败'}
}
function logout(){token='';localStorage.removeItem('ps_token');location.reload()}
function init(){if(!token)return;$('loginPage').style.display='none';$('app').style.display='block';loadSec()}
function showSec(s){sec=s;document.querySelectorAll('.tab[data-sec]').forEach(t=>t.classList.toggle('active',t.dataset.sec===s));loadSec()}
function loadSec(){
  if(sec==='articles')loadArticles();
  else if(sec==='projects')loadProjects();
  else if(sec==='categories')loadCategories();
  else if(sec==='config')loadConfig();
}

// === Articles ===
let artPage=1;
async function loadArticles(){
  const res=await fetch(API+'/admin/articles?page='+artPage,{headers:headers()});
  if(res.status===401){logout();return}
  const{data,total}=await res.json();
  const tp=Math.ceil(total/10);
  let h=`<div class="flex justify-between items-center mb-4"><h2 class="text-lg font-bold">文章管理 (${total})</h2><button onclick="editArticle(0)" class="bg-primary text-white px-4 py-2 rounded-lg text-sm hover:opacity-90">+ 新建文章</button></div>`;
  if(!data||!data.length){h+='<p class="text-center text-gray-400 py-12">暂无文章</p>'}
  else{h+=data.map(a=>`<div class="bg-white border rounded-lg p-4 mb-2 flex justify-between items-center">
    <div><div class="font-medium">${esc(a.title)} <span class="text-xs px-2 py-0.5 rounded-full ${a.status==='published'?'bg-green-50 text-green-600':'bg-yellow-50 text-yellow-600'}">${a.status==='published'?'已发布':'草稿'}</span></div>
    <div class="text-xs text-gray-400 mt-1">${a.category||'未分类'} · 👁 ${a.views} · ${(a.published_at||a.created_at||'').substring(0,10)}</div></div>
    <div class="flex gap-3 text-sm"><a class="text-primary cursor-pointer" onclick="editArticle(${a.id})">编辑</a><a class="text-red-500 cursor-pointer" onclick="deleteArticle(${a.id})">删除</a></div>
  </div>`).join('')}
  if(tp>1){h+=`<div class="flex justify-center gap-4 py-4 text-sm">${artPage>1?`<a class="text-primary cursor-pointer" onclick="artPage--;loadArticles()">← 上一页</a>`:''}<span class="text-gray-400">${artPage}/${tp}</span>${artPage<tp?`<a class="text-primary cursor-pointer" onclick="artPage++;loadArticles()">下一页 →</a>`:''}</div>`}
  $('content').innerHTML=h;
}

async function editArticle(id){
  let a={title:'',slug:'',content:'',cover:'',category_id:0,tags:'',status:'draft',summary:''};
  if(id){const res=await fetch(API+'/admin/articles/'+id,{headers:headers()});a=await res.json()}
  const cats=await(await fetch(API+'/categories')).json();
  const opts=(cats||[]).map(c=>`<option value="${c.id}" ${c.id===a.category_id?'selected':''}>${esc(c.name)}</option>`).join('');
  $('content').innerHTML=`<div class="bg-white border rounded-lg p-6">
    <h2 class="text-lg font-bold mb-4">${id?'编辑文章':'新建文章'}</h2>
    <div class="mb-3"><label class="text-xs text-gray-500">标题</label><input id="aTitle" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(a.title)}"></div>
    <div class="mb-3"><label class="text-xs text-gray-500">Slug（留空自动生成）</label><input id="aSlug" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(a.slug)}"></div>
    <div class="grid grid-cols-2 gap-3 mb-3">
      <div><label class="text-xs text-gray-500">分类</label><select id="aCat" class="w-full border rounded-lg px-3 py-2 text-sm"><option value="0">未分类</option>${opts}</select></div>
      <div><label class="text-xs text-gray-500">标签（逗号分隔）</label><input id="aTags" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(a.tags)}"></div>
    </div>
    <div class="mb-3"><label class="text-xs text-gray-500">封面图URL</label><input id="aCover" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(a.cover)}"></div>
    <div class="mb-3"><label class="text-xs text-gray-500">内容（Markdown）</label>
      <div class="grid grid-cols-2 gap-2" style="min-height:400px">
        <textarea id="aContent" class="w-full border rounded-lg px-3 py-2 text-sm font-mono resize-y" style="min-height:400px" oninput="preview()">${esc(a.content)}</textarea>
        <div id="aPreview" class="border rounded-lg px-4 py-2 overflow-auto prose prose-sm max-w-none" style="min-height:400px"></div>
      </div>
    </div>
    <div class="mb-3"><label class="text-xs text-gray-500">摘要（留空自动提取）</label><input id="aSummary" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(a.summary||'')}"></div>
    <div class="flex gap-3">
      <button onclick="saveArticle(${id},'published')" class="bg-primary text-white px-6 py-2 rounded-lg text-sm hover:opacity-90">发布</button>
      <button onclick="saveArticle(${id},'draft')" class="bg-gray-100 px-6 py-2 rounded-lg text-sm hover:bg-gray-200">存为草稿</button>
      <button onclick="loadArticles()" class="text-gray-500 px-4 py-2 text-sm">取消</button>
    </div>
  </div>`;
  preview();
}

function preview(){
  const md=$('aContent').value;
  // 简易客户端预览：发到后端渲染太慢，用简单替换
  let h=esc(md);
  h=h.replace(/^### (.+)$/gm,'<h3>$1</h3>');
  h=h.replace(/^## (.+)$/gm,'<h2>$1</h2>');
  h=h.replace(/^# (.+)$/gm,'<h1>$1</h1>');
  h=h.replace(/\*\*(.+?)\*\*/g,'<strong>$1</strong>');
  h=h.replace(/`([^`]+)`/g,'<code>$1</code>');
  h=h.replace(/\n/g,'<br>');
  $('aPreview').innerHTML=h;
}

async function saveArticle(id,status){
  const body={title:$('aTitle').value,slug:$('aSlug').value,content:$('aContent').value,summary:$('aSummary').value,cover:$('aCover').value,category_id:parseInt($('aCat').value),tags:$('aTags').value,status};
  if(!body.title){toast('标题不能为空');return}
  const url=id?API+'/admin/articles/'+id:API+'/admin/articles';
  const method=id?'PUT':'POST';
  const res=await fetch(url,{method,headers:headers(),body:JSON.stringify(body)});
  if(res.ok){toast(status==='published'?'已发布':'已保存');loadArticles()}
  else{const d=await res.json();toast(d.error||'保存失败')}
}

async function deleteArticle(id){
  if(!confirm('确定删除？'))return;
  await fetch(API+'/admin/articles/'+id,{method:'DELETE',headers:headers()});
  toast('已删除');loadArticles();
}

// === Projects ===
async function loadProjects(){
  const data=await(await fetch(API+'/projects')).json();
  let h=`<div class="flex justify-between items-center mb-4"><h2 class="text-lg font-bold">项目管理</h2><button onclick="editProject(null)" class="bg-primary text-white px-4 py-2 rounded-lg text-sm hover:opacity-90">+ 新建项目</button></div>`;
  if(!data||!data.length){h+='<p class="text-center text-gray-400 py-12">暂无项目</p>'}
  else{h+=data.map(p=>`<div class="bg-white border rounded-lg p-4 mb-2 flex justify-between items-center">
    <div><div class="font-medium">${p.icon?p.icon+' ':''}${esc(p.name)} ${!p.visible?'<span class="text-xs text-gray-400">(隐藏)</span>':''}</div>
    <div class="text-xs text-gray-400 mt-1">${esc(p.group_name||'未分组')} · ${esc(p.description).substring(0,60)}</div></div>
    <div class="flex gap-3 text-sm"><a class="text-primary cursor-pointer" onclick='editProject(${JSON.stringify(p).replace(/'/g,"&#39;")})'>编辑</a><a class="text-red-500 cursor-pointer" onclick="deleteProject(${p.id})">删除</a></div>
  </div>`).join('')}
  $('content').innerHTML=h;
}

function editProject(p){
  if(!p)p={name:'',description:'',url:'',github:'',icon:'',tags:'',group_name:'',sort_order:0,visible:true};
  $('content').innerHTML=`<div class="bg-white border rounded-lg p-6">
    <h2 class="text-lg font-bold mb-4">${p.id?'编辑项目':'新建项目'}</h2>
    <div class="grid grid-cols-2 gap-3 mb-3">
      <div><label class="text-xs text-gray-500">名称</label><input id="pName" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(p.name)}"></div>
      <div><label class="text-xs text-gray-500">图标（emoji）</label><input id="pIcon" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(p.icon)}"></div>
    </div>
    <div class="mb-3"><label class="text-xs text-gray-500">描述</label><input id="pDesc" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(p.description)}"></div>
    <div class="grid grid-cols-2 gap-3 mb-3">
      <div><label class="text-xs text-gray-500">GitHub</label><input id="pGithub" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(p.github)}"></div>
      <div><label class="text-xs text-gray-500">演示URL</label><input id="pUrl" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(p.url)}"></div>
    </div>
    <div class="grid grid-cols-3 gap-3 mb-3">
      <div><label class="text-xs text-gray-500">标签（逗号分隔）</label><input id="pTags" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(p.tags)}"></div>
      <div><label class="text-xs text-gray-500">分组</label><input id="pGroup" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(p.group_name)}" placeholder="如：开源项目"></div>
      <div><label class="text-xs text-gray-500">排序</label><input id="pSort" type="number" class="w-full border rounded-lg px-3 py-2 text-sm" value="${p.sort_order}"></div>
    </div>
    <div class="mb-4"><label class="flex items-center gap-2 text-sm"><input id="pVisible" type="checkbox" ${p.visible?'checked':''}> 在前台显示</label></div>
    <div class="flex gap-3">
      <button onclick="saveProject(${p.id||0})" class="bg-primary text-white px-6 py-2 rounded-lg text-sm hover:opacity-90">保存</button>
      <button onclick="loadProjects()" class="text-gray-500 px-4 py-2 text-sm">取消</button>
    </div>
  </div>`;
}

async function saveProject(id){
  const body={name:$('pName').value,description:$('pDesc').value,url:$('pUrl').value,github:$('pGithub').value,icon:$('pIcon').value,tags:$('pTags').value,group_name:$('pGroup').value,sort_order:parseInt($('pSort').value)||0,visible:$('pVisible').checked};
  if(!body.name){toast('名称不能为空');return}
  const url=id?API+'/admin/projects/'+id:API+'/admin/projects';
  const method=id?'PUT':'POST';
  await fetch(url,{method,headers:headers(),body:JSON.stringify(body)});
  toast('已保存');loadProjects();
}

async function deleteProject(id){
  if(!confirm('确定删除？'))return;
  await fetch(API+'/admin/projects/'+id,{method:'DELETE',headers:headers()});
  toast('已删除');loadProjects();
}

// === Categories ===
async function loadCategories(){
  const data=await(await fetch(API+'/categories')).json();
  let h=`<div class="flex justify-between items-center mb-4"><h2 class="text-lg font-bold">分类管理</h2><button onclick="editCat(0,'','')" class="bg-primary text-white px-4 py-2 rounded-lg text-sm hover:opacity-90">+ 新建分类</button></div>`;
  h+=(data||[]).map(c=>`<div class="bg-white border rounded-lg p-4 mb-2 flex justify-between items-center">
    <div><span class="font-medium">${esc(c.name)}</span> <span class="text-xs text-gray-400">/${c.slug} · ${c.count}篇</span></div>
    <div class="flex gap-3 text-sm"><a class="text-primary cursor-pointer" onclick="editCat(${c.id},'${esc(c.name)}','${esc(c.slug)}')">编辑</a><a class="text-red-500 cursor-pointer" onclick="deleteCat(${c.id})">删除</a></div>
  </div>`).join('');
  $('content').innerHTML=h;
}

function editCat(id,name,slug){
  const n=prompt('分类名称',name);if(n===null)return;
  const s=prompt('Slug',slug||n.toLowerCase());if(s===null)return;
  const url=id?API+'/admin/categories/'+id:API+'/admin/categories';
  fetch(url,{method:id?'PUT':'POST',headers:headers(),body:JSON.stringify({name:n,slug:s})}).then(()=>{toast('已保存');loadCategories()});
}

async function deleteCat(id){
  if(!confirm('删除分类？文章将变为未分类'))return;
  await fetch(API+'/admin/categories/'+id,{method:'DELETE',headers:headers()});
  toast('已删除');loadCategories();
}

// === Config ===
async function loadConfig(){
  const c=await(await fetch(API+'/config')).json();
  $('content').innerHTML=`<div class="bg-white border rounded-lg p-6">
    <h2 class="text-lg font-bold mb-4">站点设置</h2>
    <div class="grid grid-cols-2 gap-3 mb-3">
      <div><label class="text-xs text-gray-500">站点标题</label><input id="cTitle" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(c.title||'')}"></div>
      <div><label class="text-xs text-gray-500">副标题</label><input id="cSub" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(c.subtitle||'')}"></div>
    </div>
    <div class="mb-3"><label class="text-xs text-gray-500">描述</label><input id="cDesc" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(c.description||'')}"></div>
    <div class="grid grid-cols-2 gap-3 mb-3">
      <div><label class="text-xs text-gray-500">作者</label><input id="cAuthor" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(c.author||'')}"></div>
      <div><label class="text-xs text-gray-500">头像URL</label><input id="cAvatar" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(c.avatar||'')}"></div>
    </div>
    <div class="grid grid-cols-2 gap-3 mb-3">
      <div><label class="text-xs text-gray-500">GitHub</label><input id="cGithub" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(c.github||'')}"></div>
      <div><label class="text-xs text-gray-500">邮箱</label><input id="cEmail" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(c.email||'')}"></div>
    </div>
    <div class="grid grid-cols-2 gap-3 mb-3">
      <div><label class="text-xs text-gray-500">Twitter</label><input id="cTwitter" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(c.twitter||'')}"></div>
      <div><label class="text-xs text-gray-500">站点URL（用于RSS）</label><input id="cUrl" class="w-full border rounded-lg px-3 py-2 text-sm" value="${esc(c.url||'')}"></div>
    </div>
    <div class="mb-3"><label class="text-xs text-gray-500">关于页内容（Markdown）</label><textarea id="cAbout" class="w-full border rounded-lg px-3 py-2 text-sm font-mono" rows="10">${esc(c.about_markdown||'')}</textarea></div>
    <button onclick="saveConfig()" class="bg-primary text-white px-6 py-2 rounded-lg text-sm hover:opacity-90">保存</button>
  </div>`;
}

async function saveConfig(){
  const body={title:$('cTitle').value,subtitle:$('cSub').value,description:$('cDesc').value,author:$('cAuthor').value,avatar:$('cAvatar').value,github:$('cGithub').value,email:$('cEmail').value,twitter:$('cTwitter').value,url:$('cUrl').value,about_markdown:$('cAbout').value};
  await fetch(API+'/admin/config',{method:'PUT',headers:headers(),body:JSON.stringify(body)});
  toast('已保存');
}

if(token)init();
