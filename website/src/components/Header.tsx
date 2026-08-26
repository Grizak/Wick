type HeaderItem = {
  link: string;
  text: string;
};

const HeaderItems: HeaderItem[] = [
  {
    link: "https://example.com",
    text: "Example",
  },
  {
    link: "https://isaksweb.xyz",
    text: "Isaksweb.xyz",
  },
];

function HeaderItemComp({ item }: { item: HeaderItem }) {
  return (
    <>
      <li className="block rounded-2xl border-orange-500 border p-2">
        <a href={item.link}>{item.text}</a>
      </li>
    </>
  );
}

function Header() {
  return (
    <>
      <nav className="flex text-amber-300 flex-row justify-end items-center">
        <ul className="p-4">
          {HeaderItems.map((item) => {
            return <HeaderItemComp item={item} />;
          })}
        </ul>
      </nav>
    </>
  );
}

export default Header;
