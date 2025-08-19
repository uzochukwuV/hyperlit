export default function Footer() {
  return (
    <footer className="bg-white border-t border-gray-100 mt-12">
      <div className="container mx-auto px-4 py-6 flex flex-col md:flex-row items-center justify-between text-gray-600 text-sm">
        <span>© {new Date().getFullYear()} Hyperlit. All rights reserved.</span>
        <div className="flex gap-4 mt-2 md:mt-0">
          <a href="/terms" className="hover:text-primary">Terms</a>
          <a href="/privacy" className="hover:text-primary">Privacy</a>
        </div>
      </div>
    </footer>
  );
}